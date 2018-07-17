package admissions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/pravega/pravega-operator/pkg/apis/pravega/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

var server http.Server

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func admitPravegaCluster(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	pravegaCluster := v1alpha1.PravegaCluster{}

	raw := ar.Request.Object.Raw
	err := json.Unmarshal(raw, &pravegaCluster)
	if err != nil {
		logrus.Error(err)
		return toAdmissionResponse(err)
	}

	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	var msg string

	// Validate Tier2 Storage Config
	tier2Spec := pravegaCluster.Spec.Pravega.Tier2
	if tier2Spec.FileSystem == nil && tier2Spec.Ecs == nil && tier2Spec.Hdfs == nil {
		reviewResponse.Allowed = false
		msg = "No Pravega Tier2 storage specified"
	}

	if !reviewResponse.Allowed {
		reviewResponse.Result = &metav1.Status{Message: msg}
	}

	return &reviewResponse
}

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()

	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Error(err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = admit(ar)
	}

	response := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}
	// reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	resp, err := json.Marshal(response)
	if err != nil {
		glog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func servePravegaCluster(w http.ResponseWriter, r *http.Request) {
	serve(w, r, admitPravegaCluster)
}

func StartServer() {
	go start()
}

func start() {
	logrus.Infof("Starting Admissions Server")
	http.HandleFunc("/validate-pravega-cluster", servePravegaCluster)
	server := &http.Server{
		Addr: ":8443",
	}

	err := server.ListenAndServeTLS("/tls-secrets/cert.pem", "/tls-secrets/key.pem")
	if err != nil {
		panic(fmt.Sprintf("Unable to start : %v", err))
	}
}
