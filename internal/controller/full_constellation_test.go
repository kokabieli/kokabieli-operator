package controller

import (
	"context"
	"fmt"
	"time"

	kokav1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Constellation controller", func() {
	const Namespace = "test-kokabieli-constellation"
	const SecondNamespace = "test-kokabieli-constellation-second"
	const NamespaceFilteredConstellation = "test-constellation"
	const AllConstellation = "test-constellation-all"

	ctx := context.Background()

	namespacedFilteredNameTyped := types.NamespacedName{Name: NamespaceFilteredConstellation, Namespace: Namespace}
	allNameTyped := types.NamespacedName{Name: AllConstellation, Namespace: Namespace}

	createSampleConstellations := func() {
		d := "description of the constellation"
		constellation := &kokav1alpha1.Constellation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      NamespaceFilteredConstellation,
				Namespace: Namespace,
			},
			Spec: kokav1alpha1.ConstellationSpec{
				Filters:         make([]kokav1alpha1.Filter, 1),
				Name:            "Namespace filtered constellation",
				Description:     &d,
				TargetConfigMap: "test",
			},
		}
		constellation.Spec.Filters[0].Namespaces = []string{Namespace}

		err := k8sClient.Create(ctx, constellation)
		Expect(err).To(Not(HaveOccurred()))

		constellation2 := &kokav1alpha1.Constellation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      AllConstellation,
				Namespace: Namespace,
			},
			Spec: kokav1alpha1.ConstellationSpec{
				Filters:         make([]kokav1alpha1.Filter, 0),
				TargetConfigMap: "test",
				Name:            "Everything",
			},
		}

		err = k8sClient.Create(ctx, constellation2)
		Expect(err).To(Not(HaveOccurred()))
	}

	createSampleDataInterfaces := func() {
		dataInterface := &kokav1alpha1.DataInterface{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-data-interface-empty",
				Namespace: Namespace,
				Labels: map[string]string{
					"test":  "test",
					"test2": "test2",
				},
			},
			Spec: kokav1alpha1.DataInterfaceSpec{
				Name: "my-interface",
				Type: "kafka",
			},
		}

		err := k8sClient.Create(ctx, dataInterface)
		Expect(err).To(Not(HaveOccurred()))
		name := "my-interface-2"

		dataInterface2 := &kokav1alpha1.DataInterface{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-data-interface-complete",
				Namespace: Namespace,
				Labels: map[string]string{
					"test": "test",
				},
			},
			Spec: kokav1alpha1.DataInterfaceSpec{
				Name:        "my-interface-2",
				Reference:   &name,
				Type:        "kafka",
				Description: &name,
			},
		}

		err = k8sClient.Create(ctx, dataInterface2)
		Expect(err).To(Not(HaveOccurred()))
		name3 := "my-interface-3"

		dataInterface3 := &kokav1alpha1.DataInterface{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-data-interface-complete",
				Namespace: SecondNamespace,
				Labels: map[string]string{
					"test": "test",
				},
			},
			Spec: kokav1alpha1.DataInterfaceSpec{
				Name:        "my-interface-3",
				Reference:   &name3,
				Type:        "kafka",
				Description: &name3,
			},
		}

		err = k8sClient.Create(ctx, dataInterface3)
		Expect(err).To(Not(HaveOccurred()))
	}

	createSampleDataProcesses := func() {
		d := "sample interface"
		dataProcess1 := &kokav1alpha1.DataProcess{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-data-process",
				Namespace: Namespace,
				Labels: map[string]string{
					"test": "test",
				},
			},
			Spec: kokav1alpha1.DataProcessSpec{
				Name:        "my-process",
				Type:        "kafka-streams",
				Description: "my-process",
				Inputs: []kokav1alpha1.Edge{
					{Reference: Namespace + "/my-interface", Info: "hello", Trigger: true, Description: &d},
				},
				Outputs: []kokav1alpha1.Edge{
					{Reference: "my-interface-2", Info: "hello", Trigger: true, Description: &d},
				},
			},
		}
		err := k8sClient.Create(ctx, dataProcess1)
		Expect(err).To(Not(HaveOccurred()))

		dataProcess2 := &kokav1alpha1.DataProcess{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-data-process",
				Namespace: SecondNamespace,
				Labels: map[string]string{
					"test": "test",
				},
			},
			Spec: kokav1alpha1.DataProcessSpec{
				Name:        "my-process",
				Type:        "kafka-streams",
				Description: "my-process",
				Inputs: []kokav1alpha1.Edge{
					{Reference: Namespace + "/my-interface", Info: "hello", Trigger: false, Description: &d},
				},
				Outputs: []kokav1alpha1.Edge{
					{Reference: "my-interface-2", Info: "hello", Trigger: false, Description: &d},
					{Reference: "my-interface-3", Info: "hello", Trigger: true, Description: &d},
				},
			},
		}
		err = k8sClient.Create(ctx, dataProcess2)
		Expect(err).To(Not(HaveOccurred()))
	}

	createSampleDataSet := func() {
		dataSet1 := &kokav1alpha1.DataSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-data-set",
				Namespace: Namespace,
				Labels: map[string]string{
					"test":  "test",
					"test3": "test3",
				},
			},
			Spec: kokav1alpha1.DataSetSpec{
				Interfaces: []kokav1alpha1.DataInterfaceSpec{
					{
						Name: "my-interface-32",
						Type: "kafka",
						Labels: map[string]string{
							"test":  "test1",
							"test2": "test2",
						},
					},
				},
				Processes: []kokav1alpha1.DataProcessSpec{
					{
						Name:        "my-process-32",
						Type:        "kafka-streams",
						Description: "my-process",
						Inputs: []kokav1alpha1.Edge{
							{Reference: Namespace + "/my-interface", Info: "hello", Trigger: true},
						},
						Outputs: []kokav1alpha1.Edge{
							{Reference: "my-interface-2", Info: "hello", Trigger: true},
						},
					},
				},
			},
			Status: kokav1alpha1.DataSetStatus{},
		}

		Expect(k8sClient.Create(ctx, dataSet1)).To(Not(HaveOccurred()))
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Namespace,
			Namespace: Namespace,
		},
	}

	secondNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecondNamespace,
			Namespace: SecondNamespace,
		},
	}
	BeforeEach(func() {
		By("Creating the Namespaces to perform the tests")
		err := k8sClient.Create(ctx, namespace)
		Expect(err).To(Not(HaveOccurred()))
		err = k8sClient.Create(ctx, secondNamespace)
		Expect(err).To(Not(HaveOccurred()))
	})

	AfterEach(func() {
		By("Deleting all the installed crds")
		Expect(k8sClient.DeleteAllOf(ctx, &kokav1alpha1.Constellation{})).ToNot(HaveOccurred())
		Expect(k8sClient.DeleteAllOf(ctx, &kokav1alpha1.DataSet{})).ToNot(HaveOccurred())
		Expect(k8sClient.DeleteAllOf(ctx, &kokav1alpha1.DataProcess{})).ToNot(HaveOccurred())
		Expect(k8sClient.DeleteAllOf(ctx, &kokav1alpha1.DataInterface{})).ToNot(HaveOccurred())
		By("Deleting the Namespaces to perform the tests")
		err := k8sClient.Delete(ctx, namespace)
		Expect(err).ToNot(HaveOccurred())
		err = k8sClient.Delete(ctx, secondNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should successfully reconcile a custom resource for Constellation", func() {

		By("Creating the custom resource for the Kind Constellation")
		createSampleConstellations()

		By("Checking if the custom resource was successfully created")
		Eventually(func() error {
			found := &kokav1alpha1.Constellation{}
			return k8sClient.Get(ctx, namespacedFilteredNameTyped, found)
		}, time.Minute, time.Second).Should(Succeed())

		By("Checking the latest Status of the custom resource")
		Eventually(func() error {
			constellation := &kokav1alpha1.Constellation{}
			err := k8sClient.Get(ctx, namespacedFilteredNameTyped, constellation)
			if err != nil {
				return err
			}
			if constellation.Status.ConstellationResult != nil {
				if len(constellation.Status.ConstellationResult.DataInterfaceList) != 0 {
					return fmt.Errorf("data interfaces found")
				}
				if len(constellation.Status.ConstellationResult.DataProcessList) != 0 {
					return fmt.Errorf("data processes found")
				}
				return nil
			}
			return fmt.Errorf("no constellation result found")
		}, time.Second*30, time.Second).Should(Succeed())

		By("Creating the data interfaces")
		createSampleDataInterfaces()
		createSampleDataProcesses()

		By("Checking the latest Status of the custom resource")
		Eventually(func(g Gomega) error {
			constellation := &kokav1alpha1.Constellation{}
			err := k8sClient.Get(ctx, namespacedFilteredNameTyped, constellation)
			g.Expect(err).ToNot(HaveOccurred())
			if constellation.Status.ConstellationResult != nil {
				g.Expect(len(constellation.Status.ConstellationResult.DataInterfaceList)).
					To(Equal(3))
				g.Expect(constellation.Status.ConstellationResult.DataInterfaceList[2].Type).
					To(Equal("missing"))
				g.Expect(len(constellation.Status.ConstellationResult.DataProcessList)).
					To(Equal(1))
				return nil
			}
			return fmt.Errorf("no constellation result found")
		}, time.Second*30, time.Second).Should(Succeed())

		By("Checking the latest Status of the custom resource")
		Eventually(func(g Gomega) error {
			constellation := &kokav1alpha1.Constellation{}
			err := k8sClient.Get(ctx, namespacedFilteredNameTyped, constellation)
			if err != nil {

			}
			if constellation.Status.ConstellationResult != nil {
				g.Expect(len(constellation.Status.ConstellationResult.DataInterfaceList)).
					To(Equal(3))
				g.Expect(constellation.Status.ConstellationResult.DataInterfaceList[0].Type).
					To(Equal("kafka"))
				g.Expect(len(constellation.Status.ConstellationResult.DataProcessList)).
					To(Equal(1))
				return nil
			}
			return fmt.Errorf("no constellation result found")
		}, time.Second*10, time.Second).Should(Succeed())

		By("Checking the latest Status of the custom resource for all namescaces")
		Eventually(func(g Gomega) error {
			constellation := &kokav1alpha1.Constellation{}
			g.Expect(k8sClient.Get(ctx, allNameTyped, constellation)).To(Succeed())
			if constellation.Status.ConstellationResult != nil {
				g.Expect(len(constellation.Status.ConstellationResult.DataInterfaceList)).
					To(Equal(4))
				g.Expect(len(constellation.Status.ConstellationResult.DataProcessList)).
					To(Equal(2))
				return nil
			}
			return fmt.Errorf("no constellation result found")
		}, time.Second*10, time.Second).Should(Succeed())

		By("Expect the configmap to be filled")
		Eventually(func(g Gomega) error {
			configMap := &corev1.ConfigMap{}
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: "test"}, configMap)).To(Succeed())

			g.Expect(configMap.Data["index.json"]).To(Not(BeEmpty()))
			g.Expect(configMap.Data[AllConstellation+".json"]).To(Not(BeEmpty()))
			g.Expect(configMap.Data[NamespaceFilteredConstellation+".json"]).To(Not(BeEmpty()))

			return nil

		}, time.Second*10, time.Second).Should(Succeed())

		By("Loading the dataset & syncing")
		createSampleDataSet()

		By("Checking the status of our dataset that we created two objects and they exist")
		Eventually(func(g Gomega) error {
			dataSet := &kokav1alpha1.DataSet{}
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: "test-data-set"}, dataSet)).To(Succeed())

			g.Expect(len(dataSet.Status.Interfaces)).To(Equal(1))
			g.Expect(len(dataSet.Status.Processes)).To(Equal(1))

			g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: dataSet.Status.Interfaces["my-interface-32"].Name}, &kokav1alpha1.DataInterface{})).To(Succeed())
			g.Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: dataSet.Status.Processes["my-process-32"].Name}, &kokav1alpha1.DataProcess{})).To(Succeed())

			return nil

		})
	})
})
