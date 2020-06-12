/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1_test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1beta1"
)

var _ = Describe("PravegaCluster Status", func() {

	var p v1beta1.PravegaCluster
	Context("checking for default values", func() {
		BeforeEach(func() {
			p.Status.VersionHistory = nil
			p.Status.CurrentVersion = "0.5.0"
			p.Status.Init()
		})
		It("should contains pods ready condition and it is false status", func() {
			_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
			Ω(condition.Status).To(Equal(corev1.ConditionFalse))
		})
		It("should contains upgrade ready condition and it is false status", func() {
			_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
			Ω(condition.Status).To(Equal(corev1.ConditionFalse))
		})
		It("should contains pods ready condition and it is false status", func() {
			_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionError)
			Ω(condition.Status).To(Equal(corev1.ConditionFalse))
		})
	})

	Context("checking for version history", func() {
		BeforeEach(func() {
			p.Status.CurrentVersion = "0.4.4"
			p.Status.Init()
			p.Status.VersionHistory = []string{"0.6.0", "0.7.0"}
			p.Status.AddToVersionHistory("0.5.0")
		})
		It("version should get update correctly", func() {
			lastVersion := p.Status.GetLastVersion()
			Ω(lastVersion).To(Equal("0.5.0"))
		})

	})

	BeforeEach(func() {
		p = v1beta1.PravegaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("manually set pods ready condition to be true", func() {
		BeforeEach(func() {
			condition := v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionPodsReady,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			p.Status.Conditions = append(p.Status.Conditions, condition)
		})

		It("should contains pods ready condition and it is true status", func() {
			_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})
	Context("manually set pods upgrade condition to be true", func() {
		BeforeEach(func() {
			condition := v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionUpgrading,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			p.Status.Conditions = append(p.Status.Conditions, condition)
		})

		It("should contains pods upgrade condition and it is true status", func() {
			_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})
	Context("manually set pods Error condition to be true", func() {
		BeforeEach(func() {
			condition := v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionError,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			p.Status.Conditions = append(p.Status.Conditions, condition)
		})

		It("should contains pods error condition and it is true status", func() {
			_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionError)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})

	Context("set conditions", func() {
		Context("set pods ready condition to be true", func() {
			BeforeEach(func() {
				p.Status.SetPodsReadyConditionFalse()
				p.Status.SetPodsReadyConditionTrue()
			})
			It("should have pods ready condition with true status", func() {
				_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
			It("should have pods ready condition with true status using function", func() {
				Ω(p.Status.IsClusterInReadyState()).To(Equal(true))
			})
		})
		Context("set pod ready condition to be false", func() {
			BeforeEach(func() {
				p.Status.SetPodsReadyConditionTrue()
				p.Status.SetPodsReadyConditionFalse()
			})

			It("should have ready condition with false status", func() {
				_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})
			It("should have ready condition with false status using function", func() {
				Ω(p.Status.IsClusterInReadyState()).To(Equal(false))
			})
			It("should have updated timestamps", func() {
				_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
				// TODO: check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
		Context("set conditions for upgrade", func() {
			Context("set pods upgrade condition to be true", func() {
				BeforeEach(func() {
					p.Status.SetUpgradingConditionFalse()
					p.Status.SetUpgradingConditionTrue(" ", " ")
					p.Status.UpdateProgress("UpdatingControllerReason", "0")
				})
				It("should have pods upgrade condition with true status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
					Ω(condition.Status).To(Equal(corev1.ConditionTrue))
					Ω(condition.Message).To(Equal("0"))
					Ω(condition.Reason).To(Equal("UpdatingControllerReason"))
				})
				It("should have pods upgrade condition with true status using function", func() {
					Ω(p.Status.IsClusterInUpgradingState()).To(Equal(true))
				})
				It("Checking GetlastCondition function and It should return UpgradeCondition as cluster in Upgrading state", func() {
					condition := p.Status.GetLastCondition()
					Ω(string(condition.Type)).To(Equal(v1beta1.ClusterConditionUpgrading))
				})
				It("Checking ClusterInUpgradeFailedOrRollbackState should return false ", func() {
					Ω(p.Status.IsClusterInUpgradeFailedOrRollbackState()).To(Equal(false))
				})
			})

			Context("set pod upgrade condition to be false", func() {
				BeforeEach(func() {
					p.Status.SetUpgradingConditionTrue(" ", " ")
					p.Status.SetUpgradingConditionFalse()
				})

				It("should have upgrade condition with false status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
					Ω(condition.Status).To(Equal(corev1.ConditionFalse))
				})

				It("should have upgrade condition with false status using function", func() {
					Ω(p.Status.IsClusterInUpgradingState()).To(Equal(false))
				})

				It("Checking GetlastCondition function and It should return nil as not in Upgrading state", func() {
					condition := p.Status.GetLastCondition()
					Ω(condition).To(BeNil())
				})

				It("should have updated timestamps", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
					//check the timestamps
					Ω(condition.LastUpdateTime).NotTo(Equal(""))
					Ω(condition.LastTransitionTime).NotTo(Equal(""))
				})
			})
		})
		Context("set conditions for Error", func() {
			Context("set pods Error condition  upgrade failed to be true", func() {
				BeforeEach(func() {
					p.Status.SetErrorConditionFalse()
					p.Status.SetErrorConditionTrue("UpgradeFailed", " ")
				})
				It("should have pods Error condition with true status using function", func() {
					Ω(p.Status.IsClusterInUpgradeFailedState()).To(Equal(true))
				})
				It("should have pods Error condition with true status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionError)
					Ω(condition.Status).To(Equal(corev1.ConditionTrue))

				})
				It("Checking ClusterInUpgradeFailedOrRollbackState and It should return true", func() {
					Ω(p.Status.IsClusterInUpgradeFailedOrRollbackState()).To(Equal(true))
				})

			})

			Context("set pods Error condition  rollback failed to be true", func() {
				BeforeEach(func() {
					p.Status.SetErrorConditionFalse()
					p.Status.SetErrorConditionTrue("RollbackFailed", " ")
				})
				It("should return rollback failed state to true using function", func() {
					Ω(p.Status.IsClusterInRollbackFailedState()).To(Equal(true))
				})
				It("should have pods Error condition with true status using function", func() {
					Ω(p.Status.IsClusterInErrorState()).To(Equal(true))
				})
				It("should have pods Error condition with true status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionError)
					Ω(condition.Status).To(Equal(corev1.ConditionTrue))
					Ω(condition.Message).To(Equal(" "))
					Ω(condition.Reason).To(Equal("RollbackFailed"))
				})
				It("should return rollback failed state to false using function", func() {
					p.Status.SetErrorConditionTrue("some err", "")
					Ω(p.Status.IsClusterInRollbackFailedState()).To(Equal(false))
				})
			})
			Context("set pod Error condition to be false", func() {
				BeforeEach(func() {
					p.Status.SetErrorConditionTrue("UpgradeFailed", " ")
					p.Status.SetErrorConditionFalse()
				})

				It("should have Error condition with false status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionError)
					Ω(condition.Status).To(Equal(corev1.ConditionFalse))
				})

				It("should have Error condition with false status using function", func() {
					Ω(p.Status.IsClusterInUpgradeFailedState()).To(Equal(false))
				})
				It("cluster in error state should return false", func() {
					Ω(p.Status.IsClusterInErrorState()).To(Equal(false))
				})

				It("should have updated timestamps", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionError)
					//check the timestamps
					Ω(condition.LastUpdateTime).NotTo(Equal(""))
					Ω(condition.LastTransitionTime).NotTo(Equal(""))
				})
			})
		})
		Context("set conditions for rollback", func() {
			Context("set pods rollback condition to be true", func() {
				BeforeEach(func() {
					p.Status.SetRollbackConditionFalse()
					p.Status.SetRollbackConditionTrue(" ", " ")
					p.Status.UpdateProgress("UpgradeErrorReason", "")
				})
				It("should have pods rollback condition with true status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionRollback)
					Ω(condition.Status).To(Equal(corev1.ConditionTrue))
					Ω(condition.Message).To(Equal(""))
					Ω(condition.Reason).To(Equal("UpgradeErrorReason"))
				})
				It("should have pods rollback condition with true status using function", func() {
					Ω(p.Status.IsClusterInRollbackState()).To(Equal(true))
				})
				It("Checking GetlastCondition function and It should return RollbackCondition as cluster in Rollback state", func() {
					condition := p.Status.GetLastCondition()
					Ω(string(condition.Type)).To(Equal(v1beta1.ClusterConditionRollback))
				})
				It("Checking ClusterInUpgradeFailedOrRollbackState and It should return true", func() {
					Ω(p.Status.IsClusterInUpgradeFailedOrRollbackState()).To(Equal(true))
				})
			})
			Context("set pods rollback condition to be false", func() {
				BeforeEach(func() {
					p.Status.SetRollbackConditionTrue(" ", " ")
					p.Status.SetRollbackConditionFalse()
				})
				It("should have pods rollback condition with false status", func() {
					_, condition := p.Status.GetClusterCondition(v1beta1.ClusterConditionRollback)
					Ω(condition.Status).To(Equal(corev1.ConditionFalse))
				})
				It("should have pods rollback condition with false status using function", func() {
					Ω(p.Status.IsClusterInRollbackState()).To(Equal(false))
				})
				It("Checking GetlastCondition function and It should return nil as cluster not in Rollback state", func() {
					condition := p.Status.GetLastCondition()
					Ω(condition).To(BeNil())
				})
			})
		})
	})
})
