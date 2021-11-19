package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"strings"

	directcsi "github.com/minio/direct-csi/pkg/apis/direct.csi.min.io/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func toYAML(obj interface{}) (string, error) {
	formattedObj, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("error marshaling to YAML: %v", err)
	}
	return string(formattedObj), nil
}

func logYAML(obj interface{}) error {
	y, err := toYAML(obj)
	if err != nil {
		return err
	}

	fmt.Print(string(y))
	fmt.Println()
	fmt.Println("---")
	fmt.Println()
	return nil
}

func canonicalNameFromPath(val string) string {
	dr := strings.ReplaceAll(val, "/var/lib/direct-csi/devices", "")
	dr = strings.ReplaceAll(dr, "/dev/", "")
	return "/dev" + strings.ReplaceAll(dr, "-part-", "")
}

func printableBytes(value int64) string {
	if value == 0 {
		return "-"
	}
	return humanize.IBytes(uint64(value))
}

func getErrMessage(d directcsi.DirectCSIDrive) string {
	msg := ""
	for _, c := range d.Status.Conditions {
		switch c.Type {
		case string(directcsi.DirectCSIDriveConditionInitialized), string(directcsi.DirectCSIDriveConditionOwned):
			if c.Status != metav1.ConditionTrue {
				msg = c.Message
			}
		}
	}
	return msg
}

func getVolumeCount(d directcsi.DirectCSIDrive) string {
	volumes := "-"
	if len(d.Finalizers) > 1 {
		volumes = fmt.Sprintf("%v", len(d.Finalizers)-1)
	}
	return volumes
}
