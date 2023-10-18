package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ConstellationInfo struct {
	FileName     string         `json:"fileName"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	LastModified metav1.Time    `json:"lastModified"`
	Source       NamespacedName `json:"source"`
}
