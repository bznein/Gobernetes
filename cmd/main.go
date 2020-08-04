package main

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	tm.Clear()
	for {
		// access the API to list pods
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
}
