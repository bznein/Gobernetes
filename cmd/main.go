package main

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	tm "github.com/buger/goterm"
	kb "github.com/eiannone/keyboard"
)

type KeyboardInput int

const (
	Pod       KeyboardInput = 0
	Sts       KeyboardInput = 1
	ArrowUp   KeyboardInput = 2
	ArrowDown KeyboardInput = 3
	Delete    KeyboardInput = 4
	Quit      KeyboardInput = -1
)

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
	val := 0
	line := 0
	for {
		select {
		case modifier := <-inputChan:
			closeChan <- true
			switch modifier {
			case Pod, Sts:
				val = int(modifier)
				line = 0
			case ArrowUp:
				line--
			case ArrowDown:
				line++
			case Delete:
				deleteResource(clientset, line, val)
				line = 0
			case Quit:
				return
			}
			go showData(config, clientset, val, line, closeChan)
			go waitForInput(inputChan)
		}
	}
}

func listStsNamespaced(clientset *kubernetes.Clientset, namespace string) []appsv1.StatefulSet {
	sts, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	return sts.Items
}

func listPodNamespaced(clientset *kubernetes.Clientset, namespace string) []corev1.Pod {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	return pods.Items
}

func deleteResource(clientset *kubernetes.Clientset, line int, val int) {
	switch KeyboardInput(val) {
	case Pod:
		pod := listPodNamespaced(clientset, "nikolas")[line]
		clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	case Sts:
		sts := listStsNamespaced(clientset, "nikolas")[line]
		clientset.AppsV1().StatefulSets(sts.Namespace).Delete(context.TODO(), sts.Name, metav1.DeleteOptions{})
	}
}

func showData(config *rest.Config, clientset *kubernetes.Clientset, what int, line int, closeChan chan bool) {

	for {
		select {
		case <-closeChan:
			return
		default:
			switch what {
			case 0:
				readAndPrintPods(config, clientset, line)
			case 1:
				readAndPrintSts(config, clientset, line)
			}
		}
	}
}

func readAndPrintSts(config *rest.Config, clientset *kubernetes.Clientset, line int) {
	sts := listStsNamespaced(clientset, "nikolas")
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

	time.Sleep(time.Millisecond * 200)

}

func readAndPrintPods(config *rest.Config, clientset *kubernetes.Clientset, line int) {
	pods := listPodNamespaced(clientset, "nikolas")
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

	time.Sleep(time.Millisecond * 200)

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
		}
	}

}
