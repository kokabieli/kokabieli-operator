package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kokav1alpha1 "github.com/kokabieli/kokabieli-operator/api/v1alpha1"
)

var _ = Describe("Constellation controller", func() {
	const Namespace = "test-kokabieli-constellation"
	const SecondNamespace = "test-kokabieli-constellation-second"
	const NamespaceFilteredConstellation = "test-constellation"
	const AllConstellation = "test-constellation-all"

	ctx := context.Background()

	namespacedFilteredNameTyped := types.NamespacedName{Name: NamespaceFilteredConstellation, Namespace: Namespace}
	allNameTyped := types.NamespacedName{Name: AllConstellation, Namespace: Namespace}

	reconsileConstellation := func(name types.NamespacedName) error {
		constellationReconsiler := &ConstellationReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := constellationReconsiler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		return err
	}
	reconsileInterface := func(name types.NamespacedName) error {
		reconsiler := &DataInterfaceReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := reconsiler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		return err
	}
	reconsileProcess := func(name types.NamespacedName) error {
		reconsiler := &DataProcessReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := reconsiler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		return err
	}

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
		By("Deleting the Namespaces to perform the tests")
		_ = k8sClient.Delete(ctx, namespace)
		_ = k8sClient.Delete(ctx, secondNamespace)

	})

	It("should successfully reconcile a custom resource for Constellation", func() {

		By("Creating the custom resource for the Kind Constellation")
		createSampleConstellations()

		By("Checking if the custom resource was successfully created")
		Eventually(func() error {
			found := &kokav1alpha1.Constellation{}
			return k8sClient.Get(ctx, namespacedFilteredNameTyped, found)
		}, time.Minute, time.Second).Should(Succeed())

		By("Reconciling the custom resource created")
		err := reconsileConstellation(namespacedFilteredNameTyped)
		Expect(err).To(Not(HaveOccurred()))

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
			fmt.Printf("constellation.Status: %v\n", constellation.Status)
			return fmt.Errorf("no constellation result found")
		}, time.Second*10, time.Second).Should(Succeed())

		By("Creating the data interfaces")
		createSampleDataInterfaces()
		createSampleDataProcesses()

		By("Reconciling the custom resource created")
		Expect(reconsileConstellation(namespacedFilteredNameTyped)).To(Not(HaveOccurred()))
		Expect(reconsileConstellation(allNameTyped)).To(Not(HaveOccurred()))

		By("Checking the latest Status of the custom resource")
		Eventually(func() error {
			constellation := &kokav1alpha1.Constellation{}
			err := k8sClient.Get(ctx, namespacedFilteredNameTyped, constellation)
			if err != nil {

			}
			if constellation.Status.ConstellationResult != nil {
				Expect(len(constellation.Status.ConstellationResult.DataInterfaceList)).
					To(Equal(2))
				Expect(constellation.Status.ConstellationResult.DataInterfaceList[0].Type).
					To(Equal("missing"))
				Expect(len(constellation.Status.ConstellationResult.DataProcessList)).
					To(Equal(1))
				return nil
			}
			fmt.Printf("constellation.Status: %v\n", constellation.Status)
			return fmt.Errorf("no constellation result found")
		}, time.Second*10, time.Second).Should(Succeed())

		By("Full reconsiliation of the custom resources created")
		Expect(reconsileInterface(types.NamespacedName{Name: "test-data-interface-empty", Namespace: Namespace})).To(Not(HaveOccurred()))
		Expect(reconsileInterface(types.NamespacedName{Name: "test-data-interface-complete", Namespace: Namespace})).To(Not(HaveOccurred()))
		Expect(reconsileInterface(types.NamespacedName{Name: "test-data-interface-complete", Namespace: SecondNamespace})).To(Not(HaveOccurred()))
		Expect(reconsileProcess(types.NamespacedName{Name: "test-data-process", Namespace: Namespace})).To(Not(HaveOccurred()))
		Expect(reconsileProcess(types.NamespacedName{Name: "test-data-process", Namespace: SecondNamespace})).To(Not(HaveOccurred()))
		Expect(reconsileConstellation(namespacedFilteredNameTyped)).To(Not(HaveOccurred()))
		Expect(reconsileConstellation(allNameTyped)).To(Not(HaveOccurred()))

		By("Checking the latest Status of the custom resource")
		Eventually(func() error {
			constellation := &kokav1alpha1.Constellation{}
			err := k8sClient.Get(ctx, namespacedFilteredNameTyped, constellation)
			if err != nil {

			}
			if constellation.Status.ConstellationResult != nil {
				Expect(len(constellation.Status.ConstellationResult.DataInterfaceList)).
					To(Equal(3))
				Expect(constellation.Status.ConstellationResult.DataInterfaceList[0].Type).
					To(Equal("kafka"))
				Expect(len(constellation.Status.ConstellationResult.DataProcessList)).
					To(Equal(1))
				return nil
			}
			fmt.Printf("constellation.Status: %v\n", constellation.Status)
			return fmt.Errorf("no constellation result found")
		}, time.Second*10, time.Second).Should(Succeed())

		By("Checking the latest Status of the custom resource for all namescaces")
		Eventually(func() error {
			constellation := &kokav1alpha1.Constellation{}
			Expect(k8sClient.Get(ctx, allNameTyped, constellation)).To(Succeed())
			if constellation.Status.ConstellationResult != nil {
				Expect(len(constellation.Status.ConstellationResult.DataInterfaceList)).
					To(Equal(4))
				Expect(len(constellation.Status.ConstellationResult.DataProcessList)).
					To(Equal(2))
				return nil
			}
			fmt.Printf("constellation.Status: %v\n", constellation.Status)
			return fmt.Errorf("no constellation result found")
		}, time.Second*10, time.Second).Should(Succeed())

		By("Expect the configmap to be filled")
		Eventually(func() error {
			configMap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: Namespace, Name: "test"}, configMap)).To(Succeed())

			Expect(configMap.Data["index.json"]).To(Not(BeEmpty()))
			Expect(configMap.Data[AllConstellation+".json"]).To(Not(BeEmpty()))
			Expect(configMap.Data[NamespaceFilteredConstellation+".json"]).To(Not(BeEmpty()))

			return nil

		}, time.Second*10, time.Second).Should(Succeed())

	})
})
