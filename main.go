package main

import (
	"bytes"
	"log"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Import the program's configuration settings.
		cfg := config.New(ctx, "")
		machineType, err := cfg.Try("machineType")
		if err != nil {
			machineType = "f1-micro"
		}

		osImage, err := cfg.Try("osImage")
		if err != nil {
			osImage = "debian-11"
		}

		instanceTag, err := cfg.Try("instanceTag")
		if err != nil {
			instanceTag = "webserver"
		}

		instanceName, err := cfg.Try("name")
		if err != nil {
			instanceName = "instance"
		}

		type Data struct {
			TsKey string
		}

		err = godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		var b bytes.Buffer
		var keyData = Data{TsKey: os.Getenv("TSKEY")}
		t := template.Must(template.ParseFiles("startup.sh"))
		if err := t.Execute(&b, keyData); err != nil {
			log.Fatal("cannot parse tempalte")
		}

		// Create the virtual machine.
		instance, err := compute.NewInstance(ctx, instanceName, &compute.InstanceArgs{
			MachineType: pulumi.String(machineType),
			BootDisk: compute.InstanceBootDiskArgs{
				InitializeParams: compute.InstanceBootDiskInitializeParamsArgs{
					Image: pulumi.String(osImage),
				},
			},
			NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
				compute.InstanceNetworkInterfaceArgs{
					AccessConfigs: compute.InstanceNetworkInterfaceAccessConfigArray{
						compute.InstanceNetworkInterfaceAccessConfigArgs{
							// NatIp:       nil,
							// NetworkTier: nil,
						},
					},
					Subnetwork: pulumi.String(os.Getenv("SUBNET")),
				},
			},
			ServiceAccount: compute.InstanceServiceAccountArgs{
				Scopes: pulumi.ToStringArray([]string{
					"https://www.googleapis.com/auth/cloud-platform",
				}),
			},
			AllowStoppingForUpdate: pulumi.Bool(true),
			MetadataStartupScript:  pulumi.String(b.String()),
			Tags: pulumi.ToStringArray([]string{
				instanceTag,
			}),
		})
		if err != nil {
			return err
		}

		// Create a disk
		disk, err := compute.NewDisk(ctx, "builddisk", &compute.DiskArgs{
			Size: pulumi.Int(200),
		})
		if err != nil {
			return err
		}

		// Attach Disk
		_, err = compute.NewAttachedDisk(ctx, "attached-disk", &compute.AttachedDiskArgs{
			Instance: instance.SelfLink,
			Disk:     disk.SelfLink,
		})
		if err != nil {
			return err
		}

		instanceIp := instance.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

		// Export the instance's name, public IP address, and HTTP URL.
		ctx.Export("name", instance.Name)
		ctx.Export("ip", instanceIp)
		return nil
	})
}
