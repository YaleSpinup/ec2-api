package api

import (
	"context"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"
)

// getInstanceTypeCatalog fetches all instance type metadata and AZ availability,
// then merges them into an enriched catalog.
func (o *ec2Orchestrator) getInstanceTypeCatalog(ctx context.Context) ([]*InstanceTypeCatalogEntry, error) {
	log.Infof("building instance type catalog")

	types, err := o.ec2Client.DescribeAllInstanceTypes(ctx)
	if err != nil {
		return nil, err
	}

	offerings, err := o.ec2Client.ListAllInstanceTypeOfferings(ctx)
	if err != nil {
		return nil, err
	}

	// Build AZ map: instance_type_name -> [az1, az2, ...]
	azMap := make(map[string][]string, len(offerings))
	for _, offering := range offerings {
		name := aws.StringValue(offering.InstanceType)
		az := aws.StringValue(offering.Location)
		azMap[name] = append(azMap[name], az)
	}

	catalog := make([]*InstanceTypeCatalogEntry, 0, len(types))
	for _, t := range types {
		name := aws.StringValue(t.InstanceType)

		entry := &InstanceTypeCatalogEntry{
			Name:              name,
			CurrentGeneration: aws.BoolValue(t.CurrentGeneration),
			BareMetal:         aws.BoolValue(t.BareMetal),
			Burstable:         aws.BoolValue(t.BurstablePerformanceSupported),
			Category:          instanceTypeCategory(name),
			SupportedAZs:      azMap[name],
		}

		if t.VCpuInfo != nil {
			entry.VCpu = aws.Int64Value(t.VCpuInfo.DefaultVCpus)
		}

		if t.MemoryInfo != nil {
			entry.MemoryMiB = aws.Int64Value(t.MemoryInfo.SizeInMiB)
		}

		if t.ProcessorInfo != nil {
			entry.Processor = aws.StringValue(t.ProcessorInfo.Manufacturer)
			for _, arch := range t.ProcessorInfo.SupportedArchitectures {
				entry.Architectures = append(entry.Architectures, aws.StringValue(arch))
			}
		}

		if t.GpuInfo != nil && len(t.GpuInfo.Gpus) > 0 {
			gpu := t.GpuInfo.Gpus[0]
			entry.Gpu = &InstanceTypeGpuInfo{
				Count:          aws.Int64Value(gpu.Count),
				Name:           aws.StringValue(gpu.Name),
				Manufacturer:   aws.StringValue(gpu.Manufacturer),
				TotalMemoryMiB: aws.Int64Value(t.GpuInfo.TotalGpuMemoryInMiB),
			}
			if gpu.MemoryInfo != nil {
				entry.Gpu.MemoryMiB = aws.Int64Value(gpu.MemoryInfo.SizeInMiB)
			}
		}

		sort.Strings(entry.SupportedAZs)
		catalog = append(catalog, entry)
	}

	sort.Slice(catalog, func(i, j int) bool {
		return catalog[i].Name < catalog[j].Name
	})

	log.Infof("built catalog with %d instance types", len(catalog))
	return catalog, nil
}

// instanceTypeCategory derives an AWS-aligned category from the instance type name.
func instanceTypeCategory(instanceType string) string {
	if len(instanceType) == 0 {
		return "other"
	}

	// Check multi-letter families first (longer prefix match takes priority)
	switch {
	case strings.HasPrefix(instanceType, "mac"):
		return "general"
	case strings.HasPrefix(instanceType, "hpc"):
		return "hpc"
	case strings.HasPrefix(instanceType, "trn"),
		strings.HasPrefix(instanceType, "inf"),
		strings.HasPrefix(instanceType, "dl"),
		strings.HasPrefix(instanceType, "vt"):
		return "accelerated"
	case strings.HasPrefix(instanceType, "im"),
		strings.HasPrefix(instanceType, "is"):
		return "storage"
	}

	// Single-letter families
	switch instanceType[0] {
	case 't', 'm', 'a':
		return "general"
	case 'c':
		return "compute"
	case 'r', 'x', 'z', 'u':
		return "memory"
	case 'p', 'g':
		return "accelerated"
	case 'i', 'd', 'h':
		return "storage"
	default:
		return "other"
	}
}

