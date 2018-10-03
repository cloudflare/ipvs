package ipvs_test

import (
	"log"

	"github.com/cloudflare/ipvs"
)

func Example() {
	c, err := ipvs.New()
	if err != nil {
		log.Fatalf("error updating service: %v", err)
	}

	services, err := c.Services()
	if err != nil {
		log.Fatalf("error fetching services: %v", err)
	}

	for _, svc := range services {
		log.Printf("%s:%d/%s %s", svc.Address, svc.Port, svc.Protocol, svc.Scheduler)
		svc.Scheduler = "rr"

		err := c.UpdateService(svc.Service)
		if err != nil {
			log.Fatalf("error updating service: %v", err)
		}
	}
}
