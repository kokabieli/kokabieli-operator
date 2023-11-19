package controller

import (
	"context"
	"time"

	kokav1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Constellation controller", func() {
	const Namespace = "test-kokabieli-constellation"
	const SecondNamespace = "test-kokabieli-constellation-second"

	BeforeEach(func() {
		By("Creating the Namespaces to perform the tests")
		createIgnoreExisting(buildNamespace(Namespace))
		createIgnoreExisting(buildNamespace(SecondNamespace))
	})

	AfterEach(func() {
		By("Deleting all resources")
		Eventually(func(g Gomega) error {
			for _, ns := range []string{Namespace, SecondNamespace} {
				constellations := &kokav1alpha1.ConstellationList{}
				g.Expect(k8sClient.List(context.Background(), constellations, client.InNamespace(ns))).To(Succeed())
				for _, l := range constellations.Items {
					del(&l)
				}
				datasets := &kokav1alpha1.DataSetList{}
				g.Expect(k8sClient.List(context.Background(), datasets, client.InNamespace(ns))).To(Succeed())
				for _, l := range datasets.Items {
					del(&l)
				}
				dataProcesses := &kokav1alpha1.DataProcessList{}
				g.Expect(k8sClient.List(context.Background(), dataProcesses, client.InNamespace(ns))).To(Succeed())
				for _, l := range dataProcesses.Items {
					del(&l)
				}
				dataInterfaces := &kokav1alpha1.DataInterfaceList{}
				g.Expect(k8sClient.List(context.Background(), dataInterfaces, client.InNamespace(ns))).To(Succeed())
				for _, l := range dataInterfaces.Items {
					del(&l)
				}
			}
			for _, ns := range []string{Namespace, SecondNamespace} {
				constellations := &kokav1alpha1.ConstellationList{}
				g.Expect(k8sClient.List(context.Background(), constellations, client.InNamespace(ns))).To(Succeed())
				g.Expect(len(constellations.Items)).To(Equal(0))

				datasets := &kokav1alpha1.DataSetList{}
				g.Expect(k8sClient.List(context.Background(), datasets, client.InNamespace(ns))).To(Succeed())
				g.Expect(len(datasets.Items)).To(Equal(0))

				dataProcesses := &kokav1alpha1.DataProcessList{}
				g.Expect(k8sClient.List(context.Background(), dataProcesses, client.InNamespace(ns))).To(Succeed())
				g.Expect(len(dataProcesses.Items)).To(Equal(0))

				dataInterfaces := &kokav1alpha1.DataInterfaceList{}
				g.Expect(k8sClient.List(context.Background(), dataInterfaces, client.InNamespace(ns))).To(Succeed())
				g.Expect(len(dataInterfaces.Items)).To(Equal(0))
			}
			return nil
		}, 40*time.Second, time.Second).Should(Succeed())
	})

	It("should be able to build a constellation", func(ctx SpecContext) {
		By("create a constellation")
		testConstellation := buildConstellation(Namespace, "test-constellation", nil, kokav1alpha1.ConstellationSpec{
			Name:            randChars(5),
			Filters:         []kokav1alpha1.Filter{},
			Description:     sp("test"),
			TargetConfigMap: "test",
		})
		create(testConstellation)

		By("Checking if the custom resource was successfully created")
		Eventually(func(g Gomega) {
			f := buildConstellation(Namespace, "test-constellation", nil, kokav1alpha1.ConstellationSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(testConstellation.Spec.Name))
		}, 20*time.Second, time.Second).Should(Succeed())
	})

	It("should be able to build a constellation up in reverse order", func(ctx SpecContext) {
		By("create a constellation")
		testConstellation := buildConstellation(Namespace, "test-constellation", nil, kokav1alpha1.ConstellationSpec{
			Name:            randChars(5),
			Filters:         []kokav1alpha1.Filter{},
			Description:     sp("test"),
			TargetConfigMap: "test",
		})
		create(testConstellation)

		By("Checking if the custom resource was successfully created")
		Eventually(func(g Gomega) {
			f := buildConstellation(Namespace, "test-constellation", nil, kokav1alpha1.ConstellationSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(testConstellation.Spec.Name))
			g.Expect(f.Status.ConstellationResult).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult.Name).To(Equal(testConstellation.Spec.Name))
			g.Expect(f.Status.ConstellationResult.Description).To(Equal(*testConstellation.Spec.Description))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList).To(BeEmpty())
			g.Expect(f.Status.ConstellationResult.DataProcessList).To(BeEmpty())

			configMap := &corev1.ConfigMap{}
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: "test"}, configMap)).To(Succeed())

			g.Expect(configMap.Data["index.json"]).To(Not(BeEmpty()))
			g.Expect(configMap.Data["test-constellation.json"]).To(Not(BeEmpty()))

		}, 20*time.Second, time.Second).Should(Succeed())

		By("Checking if the configmap was created")
		Eventually(func(g Gomega) {
			configMap := &corev1.ConfigMap{}
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: "test"}, configMap)).To(Succeed())

			g.Expect(configMap.Data["index.json"]).To(Not(BeEmpty()))
			g.Expect(configMap.Data["test-constellation.json"]).To(Not(BeEmpty()))
		}, 20*time.Second, time.Second).Should(Succeed())

		By("create a data process without existing data interfaces")
		testDataProcess := buildDataProcess(Namespace, "test-data-process", nil, kokav1alpha1.DataProcessSpec{
			Name:        randChars(5),
			Type:        randChars(5),
			Description: "test",
			Inputs: []kokav1alpha1.Edge{
				{
					Reference:   "ref1",
					Info:        randChars(10),
					Trigger:     false,
					Description: sp(randChars(10)),
				},
				{
					Reference:   "ref2",
					Info:        randChars(10),
					Trigger:     true,
					Description: sp(randChars(10)),
				},
			},
			Outputs: []kokav1alpha1.Edge{
				{
					Reference:   "ref3",
					Info:        randChars(10),
					Trigger:     false,
					Description: sp(randChars(10)),
				},
				{
					Reference:   "ref4",
					Info:        randChars(10),
					Trigger:     true,
					Description: sp(randChars(10)),
				},
			},
		})
		create(testDataProcess)

		By("Checking if the custom resource was successfully created")
		Eventually(func(g Gomega) {
			f := buildDataProcess(Namespace, "test-data-process", nil, kokav1alpha1.DataProcessSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(testDataProcess.Spec.Name))
			g.Expect(f.Spec.Type).To(Equal(testDataProcess.Spec.Type))
			g.Expect(f.Spec.Description).To(Equal(testDataProcess.Spec.Description))
			g.Expect(f.Spec.Inputs).To(Equal(testDataProcess.Spec.Inputs))
			g.Expect(f.Spec.Outputs).To(Equal(testDataProcess.Spec.Outputs))
		}, 20*time.Second, time.Second).Should(Succeed())

		By("Verify that the constellation has been updated")
		Eventually(func(g Gomega) {
			f := buildConstellation(Namespace, "test-constellation", nil, kokav1alpha1.ConstellationSpec{})
			get(g, f)
			g.Expect(f.Status).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList).To(HaveLen(4))
			g.Expect(f.Status.ConstellationResult.DataProcessList).To(HaveLen(1))

			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Inputs[0].Reference).To(Equal("ref1"))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Inputs[1].Reference).To(Equal("ref2"))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Outputs[0].Reference).To(Equal("ref3"))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Outputs[1].Reference).To(Equal("ref4"))

			g.Expect(f.Status.ConstellationResult.DataInterfaceList[0].Type).To(Equal("missing"))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList[1].Type).To(Equal("missing"))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList[2].Type).To(Equal("missing"))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList[3].Type).To(Equal("missing"))
		}, 20*time.Second, time.Second).Should(Succeed())

		By("create a data interface")
		testDataInterface := buildDataInterface(Namespace, "test-data-interface", nil, kokav1alpha1.DataInterfaceSpec{
			Name:        randChars(5),
			Type:        randChars(5),
			Description: sp("test"),
		})
		create(testDataInterface)

		By("Checking if the custom resource was successfully created")
		Eventually(func(g Gomega) {
			f := buildDataInterface(Namespace, "test-data-interface", nil, kokav1alpha1.DataInterfaceSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(testDataInterface.Spec.Name))
			g.Expect(f.Spec.Type).To(Equal(testDataInterface.Spec.Type))
			g.Expect(f.Spec.Description).To(Equal(testDataInterface.Spec.Description))
			g.Expect(f.Status).To(Not(BeNil()))
			g.Expect(f.Status.UsedInDataProcesses).To(BeEmpty())
			g.Expect(f.Status.UsedReference).To(Equal(Namespace + "/test-data-interface"))
		}, 20*time.Second, time.Second).Should(Succeed())

		By("create the other data interfaces")
		dataInterfaceRef1 := buildDataInterface(Namespace, "test-data-interface-ref1", nil, kokav1alpha1.DataInterfaceSpec{
			Name:        "Data Interface Ref 1",
			Reference:   sp("ref1"),
			Type:        randChars(6),
			Description: sp(randChars(30)),
		})
		create(dataInterfaceRef1)
		dataInterfaceRef2 := buildDataInterface(Namespace, "test-data-interface-ref2", nil, kokav1alpha1.DataInterfaceSpec{
			Name:        "Data Interface Ref 2",
			Reference:   sp("ref2"),
			Type:        randChars(6),
			Description: sp(randChars(30)),
		})
		create(dataInterfaceRef2)
		dataInterfaceRef3 := buildDataInterface(Namespace, "test-data-interface-ref3", nil, kokav1alpha1.DataInterfaceSpec{
			Name:        "Data Interface Ref 3",
			Reference:   sp("ref3"),
			Type:        randChars(6),
			Description: sp(randChars(30)),
		})
		create(dataInterfaceRef3)
		dataInterfaceRef4 := buildDataInterface(Namespace, "test-data-interface-ref4", nil, kokav1alpha1.DataInterfaceSpec{
			Name:        "Data Interface Ref 4",
			Reference:   sp("ref4"),
			Type:        randChars(6),
			Description: sp(randChars(30)),
		})
		create(dataInterfaceRef4)

		By("Checking if the custom resource was successfully created")
		Eventually(func(g Gomega) {
			f := buildDataInterface(Namespace, "test-data-interface-ref1", nil, kokav1alpha1.DataInterfaceSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(dataInterfaceRef1.Spec.Name))
			g.Expect(f.Spec.Type).To(Equal(dataInterfaceRef1.Spec.Type))
			g.Expect(f.Spec.Description).To(Equal(dataInterfaceRef1.Spec.Description))
			g.Expect(f.Status).To(Not(BeNil()))
			g.Expect(f.Status.UsedInDataProcesses).To(HaveLen(1))
			g.Expect(f.Status.UsedInDataProcesses[0].Namespace).To(Equal(Namespace))
			g.Expect(f.Status.UsedInDataProcesses[0].Name).To(Equal("test-data-process"))
			g.Expect(f.Status.UsedReference).To(Equal("ref1"))
		}, 20*time.Second, time.Second).Should(Succeed())

		By("Checking if the constellation was updated")
		Eventually(func(g Gomega) {
			f := buildConstellation(Namespace, "test-constellation", nil, kokav1alpha1.ConstellationSpec{})
			get(g, f)
			g.Expect(f.Status).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList).To(HaveLen(5))
			g.Expect(f.Status.ConstellationResult.DataProcessList).To(HaveLen(1))

			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Type).To(Equal(testDataProcess.Spec.Type))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Inputs).To(HaveLen(2))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Outputs).To(HaveLen(2))

			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Inputs[0].Reference).To(Equal("ref1"))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Inputs[1].Reference).To(Equal("ref2"))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Outputs[0].Reference).To(Equal("ref3"))
			g.Expect(f.Status.ConstellationResult.DataProcessList[0].Outputs[1].Reference).To(Equal("ref4"))

			types := []string{}
			for _, i := range f.Status.ConstellationResult.DataInterfaceList {
				types = append(types, i.Type)
			}

			g.Expect(types).To(ContainElements(
				testDataInterface.Spec.Type,
				dataInterfaceRef1.Spec.Type,
				dataInterfaceRef2.Spec.Type,
				dataInterfaceRef3.Spec.Type,
				dataInterfaceRef4.Spec.Type))
		}, 20*time.Second, time.Second).Should(Succeed())

	})

	It("test filtering in secondary namespace", func(ctx SpecContext) {
		By("create a constellation")
		testConstellation1 := buildConstellation(Namespace, "test-constellation-1", nil, kokav1alpha1.ConstellationSpec{
			Name:            randChars(5),
			Filters:         []kokav1alpha1.Filter{},
			Description:     sp("test"),
			TargetConfigMap: "test",
		})
		create(testConstellation1)
		testConstellation2 := buildConstellation(Namespace, "test-constellation-2", nil, kokav1alpha1.ConstellationSpec{
			Name: randChars(5),
			Filters: []kokav1alpha1.Filter{{
				Namespaces: []string{SecondNamespace},
				Labels:     map[string]string{"test": "selected"},
			}},
			Description:     sp("test"),
			TargetConfigMap: "test",
		})
		create(testConstellation2)

		By("create via dataset in secondary namespace")
		dataset2 := buildDataSet(SecondNamespace, "test-dataset", map[string]string{"test": "none"}, kokav1alpha1.DataSetSpec{
			Interfaces: []kokav1alpha1.DataInterfaceSpec{
				{
					Name:        "selected.2",
					Reference:   sp("selected.2"),
					Type:        randChars(5),
					Description: sp("test"),
					Labels:      map[string]string{"test": "selected"},
				},
				{
					Name:        "unselected.2",
					Reference:   sp("unselected.2"),
					Type:        randChars(5),
					Description: sp("test"),
					Labels:      map[string]string{"test": "unselected"},
				},
			},
			Processes: []kokav1alpha1.DataProcessSpec{
				{
					Name:        "selected.2",
					Type:        randChars(5),
					Description: "test",
					Inputs: []kokav1alpha1.Edge{
						{
							Reference:   "selected.2",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
						{
							Reference:   "unselected.2",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
					},
					Outputs: []kokav1alpha1.Edge{
						{
							Reference:   "unselected.1",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
					},
					Labels: map[string]string{"test": "selected"},
				},
				{
					Name:        "unselected.2",
					Type:        randChars(5),
					Description: "test",
					Inputs: []kokav1alpha1.Edge{
						{
							Reference:   "selected.2",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
						{
							Reference:   "unselected.2",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
					},
					Outputs: []kokav1alpha1.Edge{
						{
							Reference:   "unselected.1",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
					},
					Labels: map[string]string{"test": "unselected"},
				},
			},
		})
		create(dataset2)

		By("create via dataset in primary namespace")
		dataset1 := buildDataSet(Namespace, "test-dataset", map[string]string{"test": "none"}, kokav1alpha1.DataSetSpec{
			Interfaces: []kokav1alpha1.DataInterfaceSpec{
				{
					Name:        "selected.1",
					Reference:   sp("selected.1"),
					Type:        randChars(5),
					Description: sp("test"),
					Labels:      map[string]string{"test": "selected"},
				},
				{
					Name:        "unselected.1",
					Reference:   sp("unselected.1"),
					Type:        randChars(5),
					Description: sp("test"),
					Labels:      map[string]string{"test": "unselected"},
				},
			},
			Processes: []kokav1alpha1.DataProcessSpec{
				{
					Name:        "selected.1",
					Type:        randChars(5),
					Description: "test",
					Inputs: []kokav1alpha1.Edge{
						{
							Reference:   "selected.1",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
						{
							Reference:   "unselected.1",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
					},
					Outputs: []kokav1alpha1.Edge{
						{
							Reference:   "unselected.2",
							Info:        "",
							Trigger:     false,
							Description: nil,
						},
					},
					Labels: map[string]string{"test": "selected"},
				},
			},
		})
		create(dataset1)

		By("Checking if the generic constellation took data from the dataset")
		Eventually(func(g Gomega) {
			f := buildConstellation(Namespace, "test-constellation-1", nil, kokav1alpha1.ConstellationSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(testConstellation1.Spec.Name))
			g.Expect(f.Status).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList).To(HaveLen(4))
			g.Expect(f.Status.ConstellationResult.DataProcessList).To(HaveLen(3))
		}, 20*time.Second, time.Second).Should(Succeed())

		By("Checking if the filtered constellation took data from the dataset")
		Eventually(func(g Gomega) {
			f := buildConstellation(Namespace, "test-constellation-2", nil, kokav1alpha1.ConstellationSpec{})
			get(g, f)
			g.Expect(f.Spec.Name).To(Equal(testConstellation2.Spec.Name))
			g.Expect(f.Status).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult).To(Not(BeNil()))
			g.Expect(f.Status.ConstellationResult.DataInterfaceList).To(HaveLen(3))
			g.Expect(f.Status.ConstellationResult.DataProcessList).To(HaveLen(1))
		}, 20*time.Second, time.Second).Should(Succeed())

	})
})

func get(g Gomega, obj client.Object) {
	err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(obj), obj)
	g.Expect(err).To(Not(HaveOccurred()))
}

func del(obj client.Object) {
	err := k8sClient.Delete(context.Background(), obj, client.GracePeriodSeconds(0))
	Expect(err).To(Not(HaveOccurred()))
}

func sp(in string) *string {
	return &in
}

func createIgnoreExisting(obj client.Object) {
	err := k8sClient.Create(context.Background(), obj)
	Expect(client.IgnoreAlreadyExists(err)).To(Succeed())
}

func create(obj client.Object) {
	err := k8sClient.Create(context.Background(), obj)
	Expect(err).To(Not(HaveOccurred()))
}

func buildNamespace(name string) *corev1.Namespace {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: name,
		},
	}
	return namespace
}

func buildDataInterface(namespace string, name string, labels map[string]string, spec kokav1alpha1.DataInterfaceSpec) *kokav1alpha1.DataInterface {
	dataInterface := &kokav1alpha1.DataInterface{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
	return dataInterface
}

func buildDataProcess(namespace string, name string, labels map[string]string, spec kokav1alpha1.DataProcessSpec) *kokav1alpha1.DataProcess {
	dataProcess := &kokav1alpha1.DataProcess{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
	return dataProcess
}

func buildDataSet(namespace string, name string, labels map[string]string, spec kokav1alpha1.DataSetSpec) *kokav1alpha1.DataSet {
	dataSet := &kokav1alpha1.DataSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
	return dataSet
}

func buildConstellation(namespace string, name string, labels map[string]string, spec kokav1alpha1.ConstellationSpec) *kokav1alpha1.Constellation {
	constellation := &kokav1alpha1.Constellation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
	return constellation
}
