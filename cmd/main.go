package main

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	tm "github.com/buger/goterm"
	kb "github.com/eiannone/keyboard"
	rest "k8s.io/client-go/rest"
)

type KeyboardInput int

const (
	Pod        KeyboardInput = 0
	Sts        KeyboardInput = 1
	Crds       KeyboardInput = 5
	Namespaces KeyboardInput = 6
	ArrowUp    KeyboardInput = 2
	ArrowDown  KeyboardInput = 3
	Delete     KeyboardInput = 4
	Select     KeyboardInput = 7
	Quit       KeyboardInput = -1
)

var namespace string = "default"
var sleep time.Duration = 50

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/nikolas.de-giorgis/.kube/kind")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	inputChan := make(chan KeyboardInput)
	closeChan := make(chan bool)
	go showData(config, clientset, 0, 0, closeChan)
	go waitForInput(inputChan)
	line := 0
	lastWhat := Pod
	for {
		select {
		case modifier := <-inputChan:
			closeChan <- true
			switch modifier {
			case Pod, Sts, Crds, Namespaces:
				line = 0
				lastWhat = modifier
			case ArrowUp:
				line--
			case ArrowDown:
				line++
			case Delete:
				deleteResource(config, line, lastWhat)
				line = 0
			case Select:
				selectResource(config, line, lastWhat)
			case Quit:
				return
			}
			go showData(config, clientset, lastWhat, line, closeChan)
			go waitForInput(inputChan)
		}
	}
}

func listNamespaces(clientset *kubernetes.Clientset) []corev1.Namespace {
	ns, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	return ns.Items
}
func listStsNamespaced(clientset *kubernetes.Clientset) []appsv1.StatefulSet {
	sts, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	return sts.Items
}

func listPodNamespaced(clientset *kubernetes.Clientset) []corev1.Pod {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	return pods.Items
}

func listCRDs(clientset *kubernetes.Clientset, config *rest.Config) []apiextensionsv1beta1.CustomResourceDefinition {
	apiextensionsClientSet, _ := apiextensionsclientset.NewForConfig(config)
	list, _ := apiextensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	return list.Items
}

func deleteResource(config *rest.Config, line int, val KeyboardInput) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	switch val {
	case Pod:
		pod := listPodNamespaced(clientset)[line]
		clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	case Sts:
		sts := listStsNamespaced(clientset)[line]
		clientset.AppsV1().StatefulSets(sts.Namespace).Delete(context.TODO(), sts.Name, metav1.DeleteOptions{})
	case Crds:
		crd := listCRDs(clientset, config)[line]
		apiextensionsClientSet, _ := apiextensionsclientset.NewForConfig(config)
		apiextensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(context.TODO(), crd.Name, metav1.DeleteOptions{})
	case Namespaces:
		ns := listNamespaces(clientset)[line]
		clientset.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})

	}
}

func selectResource(config *rest.Config, line int, val KeyboardInput) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}

	switch val {
	// TODO NIKOLAS implement for other reources
	case Namespaces:
		ns := listNamespaces(clientset)[line]
		namespace = ns.Name
	}
}

func showData(config *rest.Config, clientset *kubernetes.Clientset, what KeyboardInput, line int, closeChan chan bool) {
	for {
		select {
		case <-closeChan:
			return
		default:
			switch what {
			case Pod:
				readAndPrintPods(config, clientset, line)
			case Sts:
				readAndPrintSts(config, clientset, line)
			case Crds:
				readAndPrintCrds(config, clientset, line)
			case Namespaces:
				readAndPrintNamespaces(config, clientset, line)
			}
		}
	}
}

func readAndPrintCrds(config *rest.Config, clientset *kubernetes.Clientset, line int) {
	list := listCRDs(clientset, config)
	tm.Clear()
	tm.MoveCursor(1, 1)
	if line < 0 {
		line = 0
	}
	if line >= len(list) {
		line = len(list) - 1
	}
	for i, crd := range list {
		if i == line {
			tm.Printf(tm.Background(tm.Color(fmt.Sprintf("CRD: %+v\t%s\n\n", crd.Name, crd.Kind), tm.RED), tm.GREEN))
			tm.Println()
		} else {
			tm.Printf("CRD: %+v\t%s\n\n", crd.Name, crd.Kind)
		}
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(time.Millisecond * sleep)

}

func readAndPrintSts(config *rest.Config, clientset *kubernetes.Clientset, line int) {
	sts := listStsNamespaced(clientset)
	tm.Clear()
	tm.MoveCursor(1, 1)
	if line < 0 {
		line = 0
	}
	if line >= len(sts) {
		line = len(sts) - 1
	}
	for i, ss := range sts {
		if i == line {
			tm.Printf(tm.Background(tm.Color(fmt.Sprintf("Statefulset: %+v\t%s\n\n", ss.Name, ss.Kind), tm.RED), tm.GREEN))
			tm.Println()
		} else {
			tm.Printf("Statefulset: %+v\t%s\n\n", ss.Name, ss.Kind)
		}
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(time.Millisecond * sleep)

}

func readAndPrintNamespaces(config *rest.Config, clientset *kubernetes.Clientset, line int) {
	ns := listNamespaces(clientset)
	tm.Clear()
	tm.MoveCursor(1, 1)
	if line < 0 {
		line = 0
	}
	if line >= len(ns) {
		line = len(ns) - 1
	}
	for i, namespace := range ns {
		if i == line {
			tm.Printf(tm.Background(tm.Color(fmt.Sprintf("Namespace: %+v\t%s\n\n", namespace.Name, namespace.Kind), tm.RED), tm.GREEN))
			tm.Println()
		} else {
			tm.Printf("Namespace: %+v\t%s\n\n", namespace.Name, namespace.Kind)
		}
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(time.Millisecond * sleep)

}

func readAndPrintPods(config *rest.Config, clientset *kubernetes.Clientset, line int) {
	pods := listPodNamespaced(clientset)
	tm.Clear()
	tm.MoveCursor(1, 1)
	if line < 0 {
		line = 0
	}
	if line >= len(pods) {
		line = len(pods) - 1
	}
	tm.Printf("There are %d pods in the cluster\n", len(pods))
	for i, pod := range pods {
		if i == line {
			tm.Printf(tm.Background(tm.Color(fmt.Sprintf("Pod: %+v\t%s\n\n", pod.Name, pod.Status.Phase), tm.RED), tm.GREEN))
			tm.Println()
		} else {
			tm.Printf("Pod: %+v\t%s\n\n", pod.Name, pod.Status.Phase)
		}
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(time.Millisecond * sleep)

}

func waitForInput(inputChan chan KeyboardInput) {
	char, key, err := kb.GetSingleKey()
	if err != nil {
		panic(err)
	}
	if char != 0 {
		switch char {
		case '0':
			inputChan <- Pod
		case '1':
			inputChan <- Sts
		case '2':
			inputChan <- Crds
		case '3':
			inputChan <- Namespaces
		case 'q', 'Q':
			inputChan <- Quit
		case 'd', 'D':
			inputChan <- Delete
		}
	} else {
		switch key {
		case kb.KeyArrowUp:
			inputChan <- ArrowUp
		case kb.KeyArrowDown:
			inputChan <- ArrowDown
		case kb.KeyEnter:
			inputChan <- Select
		}
	}

}
