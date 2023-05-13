package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/sync"
	"google.golang.org/api/compute/v1"
)

func TestSyncX(t *testing.T) {
	const (
		proj  = "bowei-gke"
		frURL = "https://www.googleapis.com/compute/v1/projects/bowei-gke/global/forwardingRules/k8s2-fr-m8dl1fmi-default-neg-demo-ing-bowo5l43"
	)

	graph := sync.NewGraph()
	id, err := cloud.ParseResourceURL(frURL)
	if err != nil {
		t.Fatal(err)
	}
	node, err := sync.NewNodeByID(id, sync.NodeAttributes{
		Ownership: sync.OwnershipManaged,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = graph.AddNode(node)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	err = sync.TransitiveClosure(ctx, theCloud, graph)
	if err != nil {
		t.Fatal(err)
	}

	t.Error(graph.Validate())

	//t.Error(pretty.Sprint(graph))
}

/*
func xTestSync1(t *testing.T) {
	var err error
	const (
		proj          = "bowei-gke"
		networkURL    = "https://www.googleapis.com/compute/v1/projects/bowei-gke/global/networks/default"
		subnetworkURL = "https://www.googleapis.com/compute/v1/projects/bowei-gke/regions/us-central1/subnetworks/default"
	)

	ctx := context.Background()
	graph := sync.NewGraph()

	mneg := sync.NewMutableNetworkEndpointGroup(proj, meta.ZonalKey("test-neg", "us-central1-b"))
	err = mneg.Access(func(x *compute.NetworkEndpointGroup) {
		x.DefaultPort = 80
		x.Network = networkURL
		x.NetworkEndpointType = "GCE_VM_IP_PORT"
		x.Subnetwork = subnetworkURL

		x.NullFields = []string{
			"AppEngine",
			"CloudFunction",
			"CloudRun",
			"PscTargetService",
		}
		x.ForceSendFields = []string{
			"Annotations",
			"Description",
		}
	})
	if err != nil {
		t.Fatalf("neg access = %v", err)
	}

	neg, err := mneg.Freeze()
	if err != nil {
		t.Fatalf("neg freeze = %v, version = %s", err, neg.Version())
	}
	graph.AddNetworkEndpointGroup(neg, &sync.NodeAttributes{Ownership: sync.OwnershipManaged})

	fixedPort := false

	mhc := sync.NewMutableHealthCheck(proj, meta.RegionalKey("test-hc", "us-central1"))
	err = mhc.Access(func(x *compute.HealthCheck) {
		x.CheckIntervalSec = 10
		x.HealthyThreshold = 3

		if fixedPort {
			x.HttpHealthCheck = &compute.HTTPHealthCheck{
				PortSpecification: "USE_FIXED_PORT",
				RequestPath:       "/",
				ProxyHeader:       "NONE",
				Port:              80,
				NullFields: []string{ // XXX: this kind of "field cannot appear" needs to be handled.
					"PortName",
				},
				ForceSendFields: []string{
					"Host",
					"Response",
				},
			}
		} else {
			x.HttpHealthCheck = &compute.HTTPHealthCheck{
				PortSpecification: "USE_SERVING_PORT",
				RequestPath:       "/",
				ProxyHeader:       "NONE",
				NullFields: []string{ // XXX: this kind of "field cannot appear" needs to be handled.
					"Port",
					"PortName",
				},
				ForceSendFields: []string{
					"Host",
					"Response",
				},
			}
		}
		x.TimeoutSec = 5
		x.Type = "HTTP"
		x.UnhealthyThreshold = 2

		x.NullFields = []string{
			"GrpcHealthCheck",
			"Http2HealthCheck",
			"HttpsHealthCheck",
			"LogConfig",
			"SslHealthCheck",
			"TcpHealthCheck",
		}
		x.ForceSendFields = []string{
			"Description",
		}
	})
	if err != nil {
		t.Fatalf("hc access = %v", err)
	}

	hc, err := mhc.Freeze()
	if err != nil {
		t.Fatalf("hc freeze = %v, version = %s", err, neg.Version())
	}
	hcn := graph.AddHealthCheck(hc, &sync.NodeAttributes{Ownership: sync.OwnershipManaged})

	hcn.Get(ctx, theCloud)
	//hcn.GenerateLocalPlan()
	t.Log(hcn.Sprint())
	err = hcn.Sync(ctx, theCloud)
	if err != nil {
		t.Fatalf("hc sync = %v", err)
	}

	mbs := sync.NewMutableBackendService("bowei-gke", meta.RegionalKey("test-bs", "us-central1"))
	err = mbs.Access(func(x *compute.BackendService) {
		x.Backends = []*compute.Backend{
			{
				Group:              neg.ResourceID().SelfLink(meta.VersionGA),
				BalancingMode:      "RATE",
				MaxRatePerEndpoint: 100,
				CapacityScaler:     1.0,

				NullFields: []string{
					"MaxConnections",
					"MaxConnectionsPerEndpoint",
					"MaxConnectionsPerInstance",
					"MaxRate",
					"MaxRatePerInstance",
					"MaxUtilization",
					"Failover",
				},

				ForceSendFields: []string{
					"Description",
				},
			},
		}
		x.HealthChecks = []string{
			hc.ResourceID().SelfLink(meta.VersionGA),
		}
		//x.Network = networkURL
		x.Protocol = "HTTP"
		x.TimeoutSec = 100
		x.LoadBalancingScheme = "INTERNAL_MANAGED"
		x.Port = 80
		x.PortName = "http"
		x.SessionAffinity = "NONE"

		x.NullFields = []string{
			"CdnPolicy",
			"CircuitBreakers",
			"CompressionMode",
			"ConnectionTrackingPolicy",
			"ConsistentHash",
			"CustomRequestHeaders",
			"CustomResponseHeaders",
			"FailoverPolicy",
			"Iap",
			"LocalityLbPolicies",
			"LocalityLbPolicy",
			"LogConfig",
			"OutlierDetection",
			"SecuritySettings",
			"ServiceBindings",
			"Subsetting",
			// XXX
			"Network",
			// For some reason, ConnectionDraining is
			// always being sent to us, even if the value
			// inside is zero.
			"ConnectionDraining",
		}
		x.ForceSendFields = []string{
			"AffinityCookieTtlSec",
			"Description",
			"EnableCDN",
			"MaxStreamDuration",
		}
	})
	if err != nil {
		t.Fatalf("mbs access = %v", err)
	}
	bs, err := mbs.Freeze()
	if err != nil {
		t.Fatalf("bs freeze = %v, version = %s", err, neg.Version())
	}
	bsn := graph.AddBackendService(bs, &sync.NodeAttributes{Ownership: sync.OwnershipManaged})
	bsn.Get(ctx, theCloud)
	//bsn.GenerateLocalPlan()
	t.Log(bsn.Sprint())
	err = bsn.Sync(ctx, theCloud)
	if err != nil {
		t.Fatalf("bs sync = %v", err)
	}

	t.Error("X")

	/*
			panic("X")

			maddr := sync.NewMutableAddress("bowei-gke", meta.RegionalKey("gkegw-yduu-default-internal-http-bwhcuniivml7", "us-central1"))
			addr, _ := maddr.Freeze()

			mfr := sync.NewMutableForwardingRule("bowei-gke", meta.RegionalKey("gkegw-yduu-default-internal-http-2jzr7e3xclhj", "us-central1"))
			fr, _ := mfr.Freeze()

			mtp := sync.NewMutableTargetHttpProxy("bowei-gke", meta.RegionalKey("gkegw-yduu-default-internal-http-2jzr7e3xclhj", "us-central1"))
			tp, _ := mtp.Freeze()

			mum := sync.NewMutableUrlMap("bowei-gke", meta.RegionalKey("gkegw-yduu-default-internal-http-2jzr7e3xclhj", "us-central1"))
			um, _ := mum.Freeze()


			bs, _ := mbs.Freeze()

			graph.AddAddress(addr, sync.OwnershipManaged)
			graph.AddForwardingRule(fr, sync.OwnershipManaged)
			graph.AddTargetHttpProxy(tp, sync.OwnershipManaged)
			graph.AddUrlMap(um, sync.OwnershipManaged)
		for _, res := range graph.Resources() {
			err := res.Get(ctx, theCloud)
			t.Errorf("get = %v", err)
			t.Error(res.Sprint())

			err = res.GenerateLocalPlan()
			t.Errorf("plan = %v", err)
			t.Error(res.Sprint())

			// err = res.Sync(ctx, theCloud)
			// t.Errorf("sync = %v", err)
		}
}
*/

func xTestCreateInstance(t *testing.T) {
	var err error
	inst := &compute.Instance{
		CanIpForward:       false,
		DeletionProtection: false,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "https://compute.googleapis.com/compute/v1/projects/debian-cloud/zones/us-central1-b/imageFamilyViews/debian-11",
				},
				Mode: "READ_WRITE",
				Type: "PERSISTENT",
			},
		},
		MachineType: "https://www.googleapis.com/compute/v1/projects/bowei-gke/zones/us-central1-b/machineTypes/n1-standard-1",
		Metadata:    &compute.Metadata{},
		Name:        "test-instance",
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "external-nat",
						Type: "ONE_TO_ONE_NAT",
					},
				},
				Network: "https://www.googleapis.com/compute/v1/projects/bowei-gke/global/networks/default",
			},
		},
		Scheduling: &compute.Scheduling{},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: "default",
				Scopes: []string{
					"https://www.googleapis.com/auth/devstorage.read_only",
					"https://www.googleapis.com/auth/logging.write",
					"https://www.googleapis.com/auth/monitoring.write",
					"https://www.googleapis.com/auth/pubsub",
					"https://www.googleapis.com/auth/service.management.readonly",
					"https://www.googleapis.com/auth/servicecontrol",
					"https://www.googleapis.com/auth/trace.append",
				},
			},
		},
	}

	fmt.Println(inst)

	//err := theCloud.Instances().Insert(context.TODO(), meta.ZonalKey("test-instance", "us-central1-b"), inst)
	//t.Errorf("instance insert = %v", err)
	//err = theCloud.Instances().Delete(context.TODO(), meta.ZonalKey("test-instance", "us-central1-b"))
	//t.Errorf("instance delete = %v", err)

	neg := &compute.NetworkEndpointGroup{
		Name:                "test-neg",
		Network:             "https://www.googleapis.com/compute/v1/projects/bowei-gke/global/networks/default",
		NetworkEndpointType: "GCE_VM_IP_PORT",
	}
	err = theCloud.NetworkEndpointGroups().Insert(context.TODO(), meta.ZonalKey("test-neg", "us-central1-b"), neg)
	fmt.Println(neg)
	t.Errorf("neg = %v", err)

	err = theCloud.NetworkEndpointGroups().Delete(context.TODO(), meta.ZonalKey("test-neg", "us-central1-b"))
	fmt.Println(neg)
	t.Errorf("neg delete = %v", err)
}
