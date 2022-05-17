/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2eutil

import (
	goctx "context"
	"fmt"
	"strings"
	"testing"
	"time"

	bkapi "github.com/pravega/bookkeeper-operator/api/v1alpha1"
	api "github.com/pravega/pravega-operator/api/v1beta1"
	"github.com/pravega/pravega-operator/pkg/util"
	zkapi "github.com/pravega/zookeeper-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	RetryInterval        = time.Second * 5
	Timeout              = time.Second * 60
	CleanupRetryInterval = time.Second * 1
	CleanupTimeout       = time.Second * 5
	ReadyTimeout         = time.Minute * 10
	UpgradeTimeout       = time.Minute * 10
	TerminateTimeout     = time.Minute * 2
	VerificationTimeout  = time.Minute * 5
)

func InitialSetup(t *testing.T, k8client client.Client, namespace string) error {
	b := &bkapi.BookkeeperCluster{}
	b.WithDefaults()
	b.Name = "bookkeeper"
	b.Namespace = namespace
	err := DeleteBKCluster(t, k8client, b)
	if err != nil {
		return err
	}

	err = WaitForBKClusterToTerminate(t, k8client, b)
	if err != nil {
		return err
	}

	z := &zkapi.ZookeeperCluster{}
	z.WithDefaults()
	z.Name = "zookeeper"
	z.Namespace = namespace

	err = DeleteZKCluster(t, k8client, z)
	if err != nil {
		return err
	}

	err = WaitForZKClusterToTerminate(t, k8client, z)
	if err != nil {
		return err
	}

	z.WithDefaults()
	z.Spec.Persistence.VolumeReclaimPolicy = "Delete"
	z.Spec.Replicas = 1
	z.Spec.Image.PullPolicy = "IfNotPresent"
	z, err = CreateZKCluster(t, k8client, z)
	if err != nil {
		return err
	}

	err = WaitForZookeeperClusterToBecomeReady(t, k8client, z, 1)
	if err != nil {
		return err
	}

	b.WithDefaults()
	b.Name = "bookkeeper"
	b.Namespace = namespace
	b.Spec.Image.ImageSpec.PullPolicy = "IfNotPresent"
	b.Spec.Version = "0.8.0"
	b, err = CreateBKCluster(t, k8client, b)
	if err != nil {
		return err
	}
	err = WaitForBookkeeperClusterToBecomeReady(t, k8client, b, 3)
	if err != nil {
		return err
	}
	// A workaround for issue 93
	err = RestartTier2(t, k8client, namespace)
	if err != nil {
		return err
	}

	return nil
}

// CreatePravegaCluster creates a PravegaCluster CR with the desired spec
func CreatePravegaCluster(t *testing.T, k8client client.Client, p *api.PravegaCluster) (*api.PravegaCluster, error) {
	log.Printf("creating pravega cluster: %s", p.Name)
	p.Spec.Pravega.Image.PullPolicy = "IfNotPresent"
	//err := k8client.Create(goctx.TODO(), p, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout, RetryInterval: CleanupRetryInterval})
	err := k8client.Create(goctx.TODO(), p)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}
	pravega := &api.PravegaCluster{}
	err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, pravega)

	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	log.Printf("created pravega cluster: %s", pravega.Name)
	return pravega, nil
}

// CreatePravegaClusterForExternalAccess creates a PravegaCluster CR with the desired spec for ExternalAccess
func CreatePravegaClusterForExternalAccess(t *testing.T, k8client client.Client, p *api.PravegaCluster) (*api.PravegaCluster, error) {
	log.Printf("creating pravega cluster with External Access: %s", p.Name)
	p.WithDefaults()
	p.Spec.Pravega.Image.PullPolicy = "IfNotPresent"
	p.Spec.BookkeeperUri = "bookkeeper-bookie-headless:3181"
	p.Spec.ExternalAccess.Enabled = true
	p.Spec.Pravega.ControllerServiceAccountName = "pravega-components"
	p.Spec.Pravega.SegmentStoreServiceAccountName = "pravega-components"
	p.Spec.Pravega.SegmentStoreReplicas = 1
	err := k8client.Create(goctx.TODO(), p)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	pravega := &api.PravegaCluster{}
	err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, pravega)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	log.Printf("created pravega cluster: %s", pravega.Name)
	return pravega, nil
}

// CreatePravegaClusterWithTlsAuth creates a PravegaCluster CR with the desired spec for both Auth and Tls
func CreatePravegaClusterWithTlsAuth(t *testing.T, k8client client.Client, p *api.PravegaCluster) (*api.PravegaCluster, error) {
	log.Printf("creating pravega cluster with Auth and Tls: %s", p.Name)
	p.Spec.Pravega.Image.PullPolicy = "IfNotPresent"
	p.Spec.BookkeeperUri = "bookkeeper-bookie-headless:3181"
	p.Spec.Authentication.Enabled = true
	p.Spec.Authentication.PasswordAuthSecret = "password-auth"
	p.Spec.TLS.Static.ControllerSecret = "controller-tls"
	p.Spec.TLS.Static.SegmentStoreSecret = "segmentstore-tls"
	p.Spec.Pravega.Options = map[string]string{
		"pravegaservice.container.count":                                    "4",
		"pravegaservice.cache.size.max":                                     "1610612736",
		"pravegaservice.zk.connect.sessionTimeout.milliseconds":             "10000",
		"attributeIndex.readBlockSize":                                      "1048576",
		"readindex.storageRead.alignment":                                   "1048576",
		"durablelog.checkpoint.commit.count.min":                            "300",
		"bookkeeper.ack.quorum.size":                                        "3",
		"controller.security.tls.enable":                                    "true",
		"controller.security.tls.server.certificate.location":               "/etc/secret-volume/controller01.pem",
		"controller.security.tls.server.privateKey.location":                "/etc/secret-volume/controller01.key.pem",
		"controller.security.tls.trustStore.location":                       "/etc/secret-volume/tls.crt",
		"controller.security.tls.server.keyStore.location":                  "/etc/secret-volume/controller01.jks",
		"controller.security.tls.server.keyStore.pwd.location":              "/etc/secret-volume/pass-secret-tls",
		"controller.security.pwdAuthHandler.accountsDb.location":            "/etc/auth-passwd-volume/pass-secret-tls-auth.txt",
		"pravegaservice.security.tls.enable":                                "true",
		"pravegaservice.security.tls.server.certificate.location":           "/etc/secret-volume/segmentstore01.pem",
		"pravegaservice.security.tls.server.privateKey.location":            "/etc/secret-volume/segmentstore01.key.pem",
		"pravegaservice.security.tls.server.keyStore.location":              "/etc/secret-volume/segmentstore01.jks",
		"pravegaservice.security.tls.truststore.location":                   "/etc/secret-volume/tls.crt",
		"autoScale.controller.connect.security.tls.enable":                  "true",
		"autoScale.controller.connect.security.tls.truststore.location":     "/etc/secret-volume/tls.crt",
		"bookkeeper.connect.security.tls.enable":                            "true",
		"bookkeeper.connect.security.tls.trustStore.location":               "empty",
		"autoScale.controller.connect.security.tls.validateHostName.enable": "true",
		"autoScale.controller.connect.security.auth.enable":                 "true",
		"controller.security.auth.delegationToken.signingKey.basis":         "secret",
		"autoScale.security.auth.token.signingKey.basis":                    "secret",
		"pravega.client.auth.token":                                         "YWRtaW46MTExMV9hYWFh",
		"pravega.client.auth.method":                                        "Basic",
		"controller.security.tls.protocolVersion":                           "TLSv1.2",
		"pravegaservice.security.tls.protocolVersion":                       "TLSv1.2",
	}
	p.Spec.Pravega.ControllerJvmOptions = []string{"-XX:MaxDirectMemorySize=1g"}

	err := k8client.Create(goctx.TODO(), p)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	pravega := &api.PravegaCluster{}

	err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, pravega)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}

	log.Printf("created pravega cluster: %s", pravega.Name)
	return pravega, nil
}

// CreatePravegaClusterWithTls creates a PravegaCluster CR with the desired spec for tls
func CreatePravegaClusterWithTls(t *testing.T, k8client client.Client, p *api.PravegaCluster) (*api.PravegaCluster, error) {
	log.Printf("creating pravega cluster with tls: %s", p.Name)
	p.Spec.Pravega.Image.PullPolicy = "IfNotPresent"
	p.Spec.BookkeeperUri = "bookkeeper-bookie-headless:3181"
	p.Spec.TLS.Static.ControllerSecret = "controller-tls"
	p.Spec.TLS.Static.SegmentStoreSecret = "segmentstore-tls"
	p.Spec.Pravega.Options = map[string]string{
		"pravegaservice.containerCount":                                     "4",
		"pravegaservice.cacheMaxSize":                                       "1073741824",
		"pravegaservice.zkSessionTimeoutMs":                                 "10000",
		"attributeIndex.readBlockSize":                                      "1048576",
		"readIndex.storageReadAlignment":                                    "1048576",
		"durableLog.checkpointMinCommitCount":                               "300",
		"bookkeeper.bkAckQuorumSize":                                        "3",
		"controller.security.tls.enable":                                    "true",
		"controller.security.tls.server.certificate.location":               "/etc/secret-volume/controller01.pem",
		"controller.security.tls.server.privateKey.location":                "/etc/secret-volume/controller01.key.pem",
		"controller.security.tls.trustStore.location":                       "/etc/secret-volume/ca-cert",
		"controller.security.tls.server.keyStore.location":                  "/etc/secret-volume/controller01.jks",
		"controller.security.tls.server.keyStore.pwd.location":              "/etc/secret-volume/pass-secret-tls",
		"controller.security.pwdAuthHandler.accountsDb.location":            "/etc/auth-passwd-volume/pass-secret-tls-auth.txt",
		"controller.security.auth.delegationToken.ttl.seconds":              "100",
		"pravegaservice.security.tls.enable":                                "true",
		"pravegaservice.security.tls.server.certificate.location":           "/etc/secret-volume/segmentstore01.pem",
		"pravegaservice.security.tls.server.privateKey.location":            "/etc/secret-volume/segmentstore01.key.pem",
		"pravegaservice.security.tls.certificate.autoReload.enable":         "true",
		"autoScale.controller.connect.security.tls.enable":                  "true",
		"autoScale.controller.connect.security.tls.truststore.location":     "/etc/secret-volume/ca-cert",
		"bookkeeper.connect.security.tls.enable":                            "false",
		"bookkeeper.connect.security.tls.trustStore.location":               "empty",
		"autoScale.controller.connect.security.tls.validateHostName.enable": "false",
		"autoScale.controller.connect.security.auth.enable":                 "false",
		"controller.security.auth.delegationToken.signingKey.basis":         "secret",
		"autoScale.security.auth.token.signingKey.basis":                    "secret",
		"pravega.client.auth.token":                                         "YWRtaW46MTExMV9hYWFh",
		"pravega.client.auth.method":                                        "Basic",
	}
	p.Spec.Pravega.SegmentStoreJVMOptions = []string{"-Xmx2g", "-XX:MaxDirectMemorySize=2g"}
	p.Spec.Pravega.ControllerJvmOptions = []string{"-XX:MaxDirectMemorySize=1g"}
	err := k8client.Create(goctx.TODO(), p)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	pravega := &api.PravegaCluster{}

	err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, pravega)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}

	log.Printf("created pravega cluster: %s", pravega.Name)
	return pravega, nil
}

func DeletePods(t *testing.T, k8client client.Client, p *api.PravegaCluster, size int) error {

	podList := corev1.PodList{}
	listOptions := []client.ListOption{
		client.InNamespace(p.GetNamespace()),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(p.LabelsForPravegaCluster())},
	}
	err := k8client.List(goctx.TODO(), &podList, listOptions...)

	if err != nil {
		return err
	}
	pod := &corev1.Pod{}

	for i := 0; i < size; i++ {
		pod = &podList.Items[i]
		log.Printf("podnameis %v", pod.Name)
		err := k8client.Delete(goctx.TODO(), pod)
		if err != nil {
			return fmt.Errorf("failed to delete pod: %v", err)
		}
		log.Printf("deleted pravega pod: %s", pod.Name)
	}
	return nil
}

// CreateZKCluster creates a ZookeeperCluster CR with the desired spec
func CreateZKCluster(t *testing.T, k8client client.Client, z *zkapi.ZookeeperCluster) (*zkapi.ZookeeperCluster, error) {
	log.Printf("creating zookeeper cluster: %s", z.Name)
	err := k8client.Create(goctx.TODO(), z)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	zookeeper := &zkapi.ZookeeperCluster{}
	err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: z.Namespace, Name: z.Name}, zookeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	log.Printf("created zookeeper cluster: %s", z.Name)
	return zookeeper, nil
}

// CreateBKCluster creates a BookkeeperCluster CR with the desired spec
func CreateBKCluster(t *testing.T, k8client client.Client, b *bkapi.BookkeeperCluster) (*bkapi.BookkeeperCluster, error) {
	log.Printf("creating bookkeeper cluster: %s", b.Name)
	b.Spec.EnvVars = "bookkeeper-configmap"
	b.Spec.ZookeeperUri = "zookeeper-client:2181"
	b.Spec.Probes.LivenessProbe.PeriodSeconds = 10
	b.Spec.Probes.ReadinessProbe.PeriodSeconds = 10
	b.Spec.Probes.LivenessProbe.TimeoutSeconds = 15
	b.Spec.Probes.ReadinessProbe.TimeoutSeconds = 15
	err := k8client.Create(goctx.TODO(), b)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}

	bookkeeper := &bkapi.BookkeeperCluster{}
	err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, bookkeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	log.Printf("created bookkeeper cluster: %s", b.Name)
	return bookkeeper, nil
}

// DeleteBKCluster deletes the BookkeeperCluster CR specified by cluster spec
func DeleteBKCluster(t *testing.T, k8client client.Client, b *bkapi.BookkeeperCluster) error {
	log.Printf("deleting bookkeeper cluster: %s", b.Name)
	err := k8client.Delete(goctx.TODO(), b)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete CR: %v", err)
	}

	log.Printf("deleted bookkeeper cluster: %s", b.Name)
	return nil
}

// DeletePravegaCluster deletes the PravegaCluster CR specified by cluster spec
func DeletePravegaCluster(t *testing.T, k8client client.Client, p *api.PravegaCluster) error {
	log.Printf("deleting pravega cluster: %s", p.Name)
	err := k8client.Delete(goctx.TODO(), p)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete CR: %v", err)
	}

	log.Printf("deleted pravega cluster: %s", p.Name)
	return nil
}

// DeleteZKCluster deletes the ZookeeperCluster CR specified by cluster spec
func DeleteZKCluster(t *testing.T, k8client client.Client, z *zkapi.ZookeeperCluster) error {
	log.Printf("deleting zookeeper cluster: %s", z.Name)
	err := k8client.Delete(goctx.TODO(), z)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete CR: %v", err)
	}
	log.Printf("deleted zookeeper cluster: %s", z.Name)
	return nil
}

// UpdatePravegaCluster updates the PravegaCluster CR
func UpdatePravegaCluster(t *testing.T, k8client client.Client, p *api.PravegaCluster) error {
	log.Printf("updating pravega cluster: %s", p.Name)
	p.Spec.Pravega.RollbackTimeout = 10
	err := k8client.Update(goctx.TODO(), p)
	if err != nil {
		return fmt.Errorf("failed to update CR: %v", err)
	}

	log.Printf("updated pravega cluster: %s", p.Name)
	return nil
}

// GetPravegaCluster returns the latest PravegaCluster CR
func GetPravegaCluster(t *testing.T, k8client client.Client, p *api.PravegaCluster) (*api.PravegaCluster, error) {
	pravega := &api.PravegaCluster{}
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: p.Name}, pravega)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	return pravega, nil
}

// GetBKCluster returns the latest BookkeeperCluster CR
func GetBKCluster(t *testing.T, k8client client.Client, b *bkapi.BookkeeperCluster) (*bkapi.BookkeeperCluster, error) {
	bookkeeper := &bkapi.BookkeeperCluster{}
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: b.Namespace, Name: b.Name}, bookkeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	return bookkeeper, nil
}

// GetZKCluster returns the latest ZookeeperCluster CR
func GetZKCluster(t *testing.T, k8client client.Client, z *zkapi.ZookeeperCluster) (*zkapi.ZookeeperCluster, error) {
	zookeeper := &zkapi.ZookeeperCluster{}
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: z.Namespace, Name: z.Name}, zookeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain created CR: %v", err)
	}
	return zookeeper, nil
}

// WaitForPravegaClusterToBecomeReady will wait until all cluster pods are ready
func WaitForPravegaClusterToBecomeReady(t *testing.T, k8client client.Client, p *api.PravegaCluster, size int) error {
	log.Printf("waiting for cluster pods to become ready: %s", p.Name)

	err := wait.Poll(RetryInterval, ReadyTimeout, func() (done bool, err error) {
		cluster, err := GetPravegaCluster(t, k8client, p)
		if err != nil {
			return false, err
		}

		log.Printf("waiting for pods to become ready (%d/%d), pods (%v)", cluster.Status.ReadyReplicas, size, cluster.Status.Members.Ready)

		_, condition := cluster.Status.GetClusterCondition(api.ClusterConditionPodsReady)
		if condition != nil && condition.Status == corev1.ConditionTrue && cluster.Status.ReadyReplicas == int32(size) {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	log.Printf("pravega cluster ready: %s", p.Name)
	return nil
}

// WaitForBooClusterToBecomeReady will wait until all Bookkeeper cluster pods are ready
func WaitForBookkeeperClusterToBecomeReady(t *testing.T, k8client client.Client, b *bkapi.BookkeeperCluster, size int) error {
	log.Printf("waiting for cluster pods to become ready: %s", b.Name)

	err := wait.Poll(RetryInterval, ReadyTimeout, func() (done bool, err error) {
		cluster, err := GetBKCluster(t, k8client, b)
		if err != nil {
			return false, err
		}

		log.Printf("waiting for pods to become ready (%d/%d), pods (%v)", cluster.Status.ReadyReplicas, size, cluster.Status.Members.Ready)

		_, condition := cluster.Status.GetClusterCondition(bkapi.ClusterConditionPodsReady)
		if condition != nil && condition.Status == corev1.ConditionTrue && cluster.Status.ReadyReplicas == int32(size) {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	log.Printf("bookkeeper cluster ready: %s", b.Name)
	return nil
}

// WaitForZookeeperClusterToBecomeReady will wait until all zookeeper cluster pods are ready
func WaitForZookeeperClusterToBecomeReady(t *testing.T, k8client client.Client, z *zkapi.ZookeeperCluster, size int) error {
	log.Printf("waiting for cluster pods to become ready: %s", z.Name)

	err := wait.Poll(RetryInterval, ReadyTimeout, func() (done bool, err error) {
		cluster, err := GetZKCluster(t, k8client, z)
		if err != nil {
			return false, err
		}

		log.Printf("waiting for pods to become ready (%d/%d), pods (%v)", cluster.Status.ReadyReplicas, size, cluster.Status.Members.Ready)

		_, condition := cluster.Status.GetClusterCondition(zkapi.ClusterConditionPodsReady)
		if condition != nil && condition.Status == corev1.ConditionTrue && cluster.Status.ReadyReplicas == int32(size) {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	log.Printf("zookeeper cluster ready: %s", z.Name)
	return nil
}

// WaitForPravegaClusterToUpgrade will wait until all pods are upgraded
func WaitForPravegaClusterToUpgrade(t *testing.T, k8client client.Client, p *api.PravegaCluster, targetVersion string) error {
	log.Printf("waiting for cluster to upgrade: %s", p.Name)

	err := wait.Poll(RetryInterval, UpgradeTimeout, func() (done bool, err error) {
		cluster, err := GetPravegaCluster(t, k8client, p)
		if err != nil {
			return false, err
		}

		_, upgradeCondition := cluster.Status.GetClusterCondition(api.ClusterConditionUpgrading)
		_, errorCondition := cluster.Status.GetClusterCondition(api.ClusterConditionError)

		log.Printf("waiting for cluster to upgrade (upgrading: %s; error: %s)", upgradeCondition.Status, errorCondition.Status)

		if errorCondition.Status == corev1.ConditionTrue {
			return false, fmt.Errorf("failed upgrading cluster: [%s] %s", errorCondition.Reason, errorCondition.Message)
		}

		if upgradeCondition.Status == corev1.ConditionFalse && cluster.Status.CurrentVersion == targetVersion {
			// Cluster upgraded
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	log.Printf("pravega cluster upgraded: %s", p.Name)
	return nil
}

// WaitForPravegaClusterToRollback will wait until all pods have completed Rollback
func WaitForPravegaClusterToRollback(t *testing.T, k8client client.Client, p *api.PravegaCluster, targetVersion string) error {
	log.Printf("waiting for cluster to Rollback: %s", p.Name)

	err := wait.Poll(RetryInterval, UpgradeTimeout, func() (done bool, err error) {
		cluster, err := GetPravegaCluster(t, k8client, p)
		if err != nil {
			return false, err
		}

		_, upgradeCondition := cluster.Status.GetClusterCondition(api.ClusterConditionRollback)
		_, errorCondition := cluster.Status.GetClusterCondition(api.ClusterConditionError)

		log.Printf("waiting for cluster to Rollback (upgrading: %s; error: %s)", upgradeCondition.Status, errorCondition.Status)

		if upgradeCondition.Status == corev1.ConditionFalse && cluster.Status.CurrentVersion == targetVersion {
			// Cluster upgraded
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	log.Printf("pravega cluster Completed Rollback: %s", p.Name)
	return nil
}

// WaitForPravegaClusterToFailUpgrade will wait till Upgrade Fails
func WaitForPravegaClusterToFailUpgrade(t *testing.T, k8client client.Client, p *api.PravegaCluster, targetVersion string) error {
	log.Printf("waiting for cluster to Fail upgrade: %s", p.Name)

	err := wait.Poll(RetryInterval, UpgradeTimeout, func() (done bool, err error) {
		cluster, err := GetPravegaCluster(t, k8client, p)
		if err != nil {
			return false, err
		}

		_, upgradeCondition := cluster.Status.GetClusterCondition(api.ClusterConditionUpgrading)
		_, errorCondition := cluster.Status.GetClusterCondition(api.ClusterConditionError)

		log.Printf("waiting for cluster to upgrade (upgrading: %s; error: %s)", upgradeCondition.Status, errorCondition.Status)

		if upgradeCondition.Status == corev1.ConditionFalse && errorCondition.Status == corev1.ConditionTrue {
			// Cluster upgraded Failed
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	log.Printf("pravega cluster upgraded: %s", p.Name)
	return nil
}

func WaitForCMPravegaClusterToUpgrade(t *testing.T, k8client client.Client, p *api.PravegaCluster) error {
	log.Printf("waiting for cluster to upgrade post cm changes: %s", p.Name)

	podList := corev1.PodList{}
	listOptions := []client.ListOption{
		client.InNamespace(p.GetNamespace()),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(p.LabelsForPravegaCluster())},
	}
	err := k8client.List(goctx.TODO(), &podList, listOptions...)

	// Checking if all pods are getting restarted

	if err != nil {
		return err
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		name := pod.Name
		log.Printf("waiting for pods to terminate, running pods (%v)", pod.Name)
		err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: name}, pod)
		start := time.Now()
		for util.IsPodReady(pod) {
			if time.Since(start) > 5*time.Minute {
				return fmt.Errorf("failed to delete Segmentstore pod (%s) for 5 mins ", name)
			}
			err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: name}, pod)
		}
	}

	//Checking if all pods are in ready state
	err = k8client.List(goctx.TODO(), &podList, listOptions...)
	if err != nil {
		return err
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		name := pod.Name
		log.Printf("waiting for pods to terminate, running pods (%v)", pod.Name)
		err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: name}, pod)
		start := time.Now()
		for !util.IsPodReady(pod) {
			if time.Since(start) > 5*time.Minute {
				return fmt.Errorf("failed to delete Segmentstore pod (%s) for 5 mins ", name)
			}
			err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: p.Namespace, Name: name}, pod)
		}
	}

	return nil
}

// WaitForPravegaClusterToTerminate will wait until all cluster pods are terminated
func WaitForPravegaClusterToTerminate(t *testing.T, k8client client.Client, p *api.PravegaCluster) error {
	log.Printf("waiting for pravega cluster to terminate: %s", p.Name)

	listOptions := []client.ListOption{
		client.InNamespace(p.GetNamespace()),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(p.LabelsForPravegaCluster())},
	}

	// Wait for Pods to terminate
	err := wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		podList := corev1.PodList{}
		err = k8client.List(goctx.TODO(), &podList, listOptions...)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}
		log.Printf("waiting for pods to terminate, running pods (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	// Wait for PVCs to terminate
	err = wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		pvcList := corev1.PersistentVolumeClaimList{}
		err = k8client.List(goctx.TODO(), &pvcList, listOptions...)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range pvcList.Items {
			pvc := &pvcList.Items[i]
			names = append(names, pvc.Name)
		}
		log.Printf("waiting for pvc to terminate (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	log.Printf("pravega cluster terminated: %s", p.Name)
	return nil
}

// WaitForZKClusterToTerminate will wait until all zookeeper cluster pods are terminated
func WaitForZKClusterToTerminate(t *testing.T, k8client client.Client, z *zkapi.ZookeeperCluster) error {
	log.Printf("waiting for zookeeper cluster to terminate: %s", z.Name)

	listOptions := []client.ListOption{
		client.InNamespace(z.GetNamespace()),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(map[string]string{"app": z.GetName()})},
	}

	// Wait for Pods to terminate
	err := wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		podList := corev1.PodList{}
		err = k8client.List(goctx.TODO(), &podList, listOptions...)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}
		log.Printf("waiting for pods to terminate, running pods (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	// Wait for PVCs to terminate
	err = wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		pvcList := corev1.PersistentVolumeClaimList{}
		err = k8client.List(goctx.TODO(), &pvcList, listOptions...)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range pvcList.Items {
			pvc := &pvcList.Items[i]
			names = append(names, pvc.Name)
		}
		log.Printf("waiting for pvc to terminate (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil

	})

	if err != nil {
		return err
	}

	log.Printf("zookeeper cluster terminated: %s", z.Name)
	return nil
}

// WaitForBKClusterToTerminate will wait until all Bookkeeper cluster pods are terminated
func WaitForBKClusterToTerminate(t *testing.T, k8client client.Client, b *bkapi.BookkeeperCluster) error {
	log.Printf("waiting for Bookkeeper cluster to terminate: %s", b.Name)

	listOptions := []client.ListOption{
		client.InNamespace(b.GetNamespace()),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(map[string]string{"bookkeeper_cluster": b.GetName()})},
	}

	// Wait for Pods to terminate
	err := wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		podList := corev1.PodList{}
		err = k8client.List(goctx.TODO(), &podList, listOptions...)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range podList.Items {
			pod := &podList.Items[i]
			names = append(names, pod.Name)
		}
		log.Printf("waiting for pods to terminate, running pods (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	// Wait for PVCs to terminate
	err = wait.Poll(RetryInterval, TerminateTimeout, func() (done bool, err error) {
		pvcList := corev1.PersistentVolumeClaimList{}
		err = k8client.List(goctx.TODO(), &pvcList, listOptions...)
		if err != nil {
			return false, err
		}

		var names []string
		for i := range pvcList.Items {
			pvc := &pvcList.Items[i]
			names = append(names, pvc.Name)
		}
		log.Printf("waiting for pvc to terminate (%v)", names)
		if len(names) != 0 {
			return false, nil
		}
		return true, nil

	})

	if err != nil {
		return err
	}

	log.Printf("bookkeeper cluster terminated: %s", b.Name)
	return nil
}

// WriteAndReadData writes sample data and reads it back from the given Pravega cluster
func WriteAndReadData(t *testing.T, k8client client.Client, p *api.PravegaCluster) error {
	log.Printf("writing and reading data from pravega cluster: %s", p.Name)
	testJob := NewTestWriteReadJob(p.Namespace, p.ServiceNameForController())
	err := k8client.Create(goctx.TODO(), testJob)

	if err != nil {
		return fmt.Errorf("failed to create job: %s", err)
	}

	err = wait.Poll(RetryInterval, VerificationTimeout, func() (done bool, err error) {
		err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: testJob.Namespace, Name: testJob.Name}, testJob)
		if err != nil {
			return false, err
		}
		if testJob.Status.CompletionTime.IsZero() {
			return false, nil
		}
		if testJob.Status.Failed > 0 {
			return false, fmt.Errorf("failed to write and read data from cluster")
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	log.Printf("pravega cluster validated: %s", p.Name)
	return nil
}

// UpdatePravegaClusterRollback updates the PravegaCluster CR for Rollback
func UpdatePravegaClusterRollback(t *testing.T, k8client client.Client, p *api.PravegaCluster) error {
	log.Printf("updating pravega cluster: %s", p.Name)
	p.Spec.Pravega.RollbackTimeout = 1
	err := k8client.Update(goctx.TODO(), p)
	if err != nil {
		return fmt.Errorf("failed to Rollback CR: %v", err)
	}

	log.Printf("completed Rollback of pravega cluster: %s", p.Name)
	return nil
}

// CheckExternalAccesss Checks if External Access is enabled or not
func CheckExternalAccesss(t *testing.T, k8client client.Client, pravega *api.PravegaCluster) error {

	ssSvc := &corev1.Service{}
	conSvc := &corev1.Service{}
	_ = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: pravega.Namespace, Name: pravega.ServiceNameForSegmentStore(0)}, ssSvc)
	_ = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: pravega.Namespace, Name: pravega.ServiceNameForController()}, conSvc)

	if len(conSvc.Status.LoadBalancer.Ingress) == 0 || len(ssSvc.Status.LoadBalancer.Ingress) == 0 {
		return fmt.Errorf("External Access is not enabled")
	}
	log.Printf("pravega cluster External Acess Validated: %s", pravega.Name)
	return nil
}

func CheckServiceExists(t *testing.T, k8client client.Client, pravega *api.PravegaCluster, svcName string) error {
	svc := &corev1.Service{}
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: pravega.Namespace, Name: svcName}, svc)
	if err != nil {
		return fmt.Errorf("service doesnt exist: %v", err)
	}
	return nil
}
func CheckStsExists(t *testing.T, k8client client.Client, pravega *api.PravegaCluster, stsName string) error {
	sts := &appsv1.StatefulSet{}
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: pravega.Namespace, Name: stsName}, sts)
	if err != nil {
		return fmt.Errorf("sts doesnt exist: %v", err)
	}

	return nil
}

func CheckConfigMapUpdated(t *testing.T, k8client client.Client, pravega *api.PravegaCluster, cmName string, key string, values []string) error {
	cm := &corev1.ConfigMap{}
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: pravega.Namespace, Name: cmName}, cm)
	if err != nil {
		return fmt.Errorf("failed to obtain configmap: %v", err)
	}
	if cm != nil {
		optvalue := cm.Data[key]
		for _, value := range values {
			if !strings.Contains(optvalue, value) {
				return fmt.Errorf("config map is not updated")
			}
		}
	} else {
		return fmt.Errorf("config map is empty")
	}
	return nil
}

// GetSts returns the sts
func GetSts(t *testing.T, k8client client.Client, stsName string) (*appsv1.StatefulSet, error) {
	sts := &appsv1.StatefulSet{}
	ns := "default"
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: ns, Name: stsName}, sts)
	if err != nil {
		return nil, fmt.Errorf("sts doesnt exist: %v", err)
	}

	return sts, nil
}

// GetDeployment returns the deployment
func GetDeployment(t *testing.T, k8client client.Client, deployName string) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}
	ns := "default"
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: ns, Name: deployName}, deploy)
	if err != nil {
		return nil, fmt.Errorf("Deployment doesnt exist: %v", err)
	}

	return deploy, nil
}

func RestartTier2(t *testing.T, k8client client.Client, namespace string) error {
	t.Log("restarting tier2 storage")
	tier2 := NewTier2(namespace)
	err := k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: tier2.Name}, tier2)
	//	_, err := f.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(goctx.TODO(), tier2.Name, metav1.GetOptions{})

	if err == nil {
		err := k8client.Delete(goctx.TODO(), tier2)
		if err != nil {
			return fmt.Errorf("failed to delete tier2: %v", err)
		}
	}

	err = wait.Poll(RetryInterval, 3*time.Minute, func() (done bool, err error) {
		err = k8client.Get(goctx.TODO(), types.NamespacedName{Namespace: namespace, Name: tier2.Name}, tier2)
		//	_, err = f.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(goctx.TODO(), tier2.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for tier2 termination: %s", err)
	}

	tier2 = NewTier2(namespace)
	err = k8client.Create(goctx.TODO(), tier2)
	if err != nil {
		return fmt.Errorf("failed to create tier2: %s", err)
	}

	log.Printf("pravega cluster tier2 restarted")
	return nil
}
