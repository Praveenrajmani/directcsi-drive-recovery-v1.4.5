package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	directcsi "github.com/minio/direct-csi/pkg/apis/direct.csi.min.io/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("invalid usage. Please check '%s --help'", args[0])
		os.Exit(1)
	}

	if len(args) == 2 && args[1] == "--help" {
		fmt.Printf("%s create|delete drive-objects-path timestamp-in-RFC3339 [--yaml]. Eg: %s create drives.yaml 2021-11-18T08:29:00Z [--yaml]", args[0], args[0])
		os.Exit(0)
	}

	if args[1] != "create" && args[1] != "delete" {
		fmt.Printf("invalid usage. Please check '%s --help'", args[0])
		os.Exit(0)
	}
	mode := args[1]

	yamlPrint := false
	if len(args) == 5 && args[4] == "--yaml" {
		yamlPrint = true
	}

	driveFilePath := args[2]
	driveBytes, err := ioutil.ReadFile(driveFilePath)
	if err != nil {
		fmt.Printf("Failed to read the file: %v", err)
		os.Exit(1)
	}

	driveList := &directcsi.DirectCSIDriveList{}
	if err := yaml.Unmarshal(driveBytes, driveList); err != nil {
		fmt.Printf("Error while unmarshalling: %v", err)
		os.Exit(1)
	}

	headers := []interface{}{
		"DRIVE",
		"CAPACITY",
		"ALLOCATED",
		"FILESYSTEM",
		"VOLUMES",
		"NODE",
		"CREATION_TS",
		"DELETETION_TS",
		"",
	}

	text.DisableColors()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row(headers))

	style := table.StyleColoredDark
	style.Color.IndexColumn = text.Colors{text.FgHiBlue, text.BgHiBlack}
	style.Color.Header = text.Colors{text.FgHiBlue, text.BgHiBlack}
	t.SetStyle(style)

	tsArg := args[3]
	ts, err := time.Parse(time.RFC3339, tsArg)
	if err != nil {
		fmt.Printf("invalid ts value: %v. please check '%s --help'", err, args[0])
		os.Exit(1)
	}
	thresholdTime := metav1.NewTime(ts)
	for _, driveItem := range driveList.Items {
		deletionTs := driveItem.GetDeletionTimestamp()
		if !deletionTs.IsZero() && driveItem.Status.DriveStatus == directcsi.DriveStatusTerminating && !deletionTs.Before(&thresholdTime) {
			if yamlPrint {
				if mode == "create" {
					driveItem.ObjectMeta.DeletionTimestamp = nil
					driveItem.ObjectMeta.DeletionGracePeriodSeconds = nil
					driveItem.Status.DriveStatus = directcsi.DriveStatusInUse
				} else {
					driveItem.ObjectMeta.Finalizers = nil
				}
				if err := logYAML(driveItem); err != nil {
					fmt.Printf("error while printing yaml: %v", err)
					os.Exit(1)
				}
			} else {
				output := []interface{}{
					canonicalNameFromPath(driveItem.Status.Path),
					printableBytes(driveItem.Status.TotalCapacity),
					printableBytes(driveItem.Status.AllocatedCapacity),
					driveItem.Status.Filesystem,
					getVolumeCount(driveItem),
					driveItem.Status.NodeName,
					driveItem.GetCreationTimestamp().String(),
					driveItem.GetDeletionTimestamp().String(),
					getErrMessage(driveItem),
				}
				t.AppendRow(output)
			}
		}
	}

	if !yamlPrint {
		t.Render()
	}
}
