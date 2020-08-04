package main

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	tm "github.com/buger/goterm"
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
	inputChan := make(chan int)
	closeChan := make(chan bool)
	go showData(config, clientset, 0, closeChan)
	go waitForInput(inputChan)
	val := 0
	for {
		select {
		case val = <-inputChan:
			closeChan <- true
			fmt.Fprintf(os.Stderr, "Changing format")
			go showData(config, clientset, val, closeChan)
			go waitForInput(inputChan)
		}
	}
}

func showData(config *rest.Config, clientset *kubernetes.Clientset, what int, closeChan chan bool) {

	for {
		select {
		case <-closeChan:
			return
		default:
			switch what {
			case 0:
				readAndPrintPods(config, clientset)
			case 1:
				readAndPrintSts(config, clientset)
			}
		}
	}
}

func readAndPrintSts(config *rest.Config, clientset *kubernetes.Clientset) {

	tm.Clear()
	sts, err := clientset.AppsV1().StatefulSets("nikolas").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	tm.MoveCursor(1, 1)

	for _, ss := range sts.Items {
		tm.Printf("Statefulset: %+v\t%s\n\n", ss.Name, ss.Kind)
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(time.Second)

}

func readAndPrintPods(config *rest.Config, clientset *kubernetes.Clientset) {

	tm.Clear()
	pods, err := clientset.CoreV1().Pods("nikolas").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
	tm.MoveCursor(1, 1)

	tm.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		tm.Printf("Pod: %+v\t%s\n\n", pod.Name, pod.Status.Phase)
	}

	tm.Flush() // Call it every time at the end of rendering

	time.Sleep(time.Second)

}

func waitForInput(inputChan chan int) {
	var i int
	_, _ = fmt.Scanf("%d", &i)
	inputChan <- i

}
