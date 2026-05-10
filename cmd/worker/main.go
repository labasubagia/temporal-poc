package main

import (
	"log"
	"os"

	"github.com/labasubagia/temporal-poc/activities"
	"github.com/labasubagia/temporal-poc/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/sysinfo"
	"go.temporal.io/sdk/worker"
)

func main() {
	temporalAddr := os.Getenv("TEMPORAL_HOST_URL")
	if temporalAddr == "" {
		temporalAddr = "localhost:7233"
	}

	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	c, err := client.NewLazyClient(client.Options{
		HostPort:  temporalAddr,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer c.Close()

	w := worker.New(c, "payment-worker", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
		SysInfoProvider: sysinfo.SysInfoProvider(),
	})


	w.RegisterWorkflow(workflow.PaymentWorkflow)
	w.RegisterWorkflow(workflow.OrderFulfillmentWorkflow)
	w.RegisterWorkflow(workflow.FailingWorkflow)
	w.RegisterWorkflow(workflow.PurchaseOrderWorkflow)
	w.RegisterActivity(activities.New())
	w.RegisterActivity(activities.NewOrderActivities())
	w.RegisterActivity(activities.NewFailingActivities())
	w.RegisterActivity(activities.NewPurchase())

	log.Println("Worker started, connecting to:", temporalAddr)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker error: %v", err)
	}

	c.Close()
	log.Println("Worker stopped")
}