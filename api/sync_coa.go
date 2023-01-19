package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func UpdateSnapshotTags(ctx context.Context, orch *ec2Orchestrator, snaps []*ec2.Snapshot) (map[string][]string, error) {
	volIds := make([]string, len(snaps))
	for i, s := range snaps {
		volIds[i] = aws.StringValue(s.VolumeId)
	}
	volIdTagMap, err := getVolumeTags(ctx, orch, volIds)
	if err != nil {
		return nil, err
	}
	stats := make(map[string][]string)
	count := 0
	for _, snap := range snaps {
		if tags := volIdTagMap[aws.StringValue(snap.VolumeId)]; len(tags) > 0 {
			count++
			if count%20 == 0 {
				fmt.Println("Sleeping for 2 secs")
				time.Sleep(10 * time.Second)
			}
			input := ec2.CreateTagsInput{
				Resources: []*string{snap.SnapshotId},
				Tags:      tags,
			}
			log.Printf("creating tag for %s, count %d\n", *snap.SnapshotId, count)
			if _, err := orch.ec2Client.Service.CreateTagsWithContext(ctx, &input); err != nil {
				log.Printf("error creating tag for %s, err %v\n", *snap.SnapshotId, err)
				stats["error"] = append(stats["error"], aws.StringValue(snap.SnapshotId))
			} else {
				stats["updated-tags"] = append(stats["updated-tags"], aws.StringValue(snap.SnapshotId))
			}
		} else {
			stats["no-tags"] = append(stats["no-tags"], aws.StringValue(snap.SnapshotId)+":"+aws.StringValue(snap.VolumeId))
		}
	}
	return stats, nil

}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getVolumeTags(ctx context.Context, orch *ec2Orchestrator, volIds []string) (map[string][]*ec2.Tag, error) {

	volIdTagMap := make(map[string][]*ec2.Tag)
	for i := 0; i < len(volIds); i += 200 {
		n := min(i+200, len(volIds))
		log.Printf("Getting ids %d:%d\n", i, n)
		vols, err := orch.getVolumes(ctx, volIds[i:n]...)
		if err != nil {
			return nil, err
		}
		log.Printf("Got ids %d\n", len(vols))
		for _, vol := range vols {
			volIdTagMap[aws.StringValue(vol.VolumeId)] = vol.Tags
		}
	}
	return volIdTagMap, nil
}

var coaFilter = &ec2.Filter{
	Name:   aws.String("tag-key"),
	Values: []*string{aws.String("ChargingAccount")},
}

func SnapshotsWithoutCOA(ctx context.Context, orch *ec2Orchestrator) ([]*ec2.Snapshot, error) {
	allSnapshots, _, err := orch.listSnapshots(ctx, 0, nil)
	if err != nil {
		return nil, err
	}
	snapshotsWithCOA, _, err := orch.listSnapshots(ctx, 0, nil, coaFilter)
	if err != nil {
		return nil, err
	}
	return subtract(allSnapshots, snapshotsWithCOA), nil
}

func subtract(a, b []*ec2.Snapshot) []*ec2.Snapshot {
	var newList []*ec2.Snapshot
	aSet, bSet := createSet(a), createSet(b)
	for key, val := range aSet {
		if _, ok := bSet[key]; !ok {
			newList = append(newList, val)
		}
	}
	return newList
}

func createSet(ss []*ec2.Snapshot) map[string]*ec2.Snapshot {
	set := make(map[string]*ec2.Snapshot, len(ss))
	for _, s := range ss {
		set[aws.StringValue(s.SnapshotId)] = s
	}
	return set
}
