package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/google/go-cmp/cmp"
)

const (
	// StudentsTableName is "students"
	StudentsTableName = "students"
	// ScoresTableName is "scores"
	ScoresTableName = "scores"
)

func loadTestDataFromJSON(ctx context.Context, client *bigquery.Client, tableName, jsonFilePath string) error {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return err
	}
	source := bigquery.NewReaderSource(jsonFile)
	source.SourceFormat = bigquery.JSON
	loader := client.Dataset(dataset).Table(tableName).LoaderFrom(source)
	loader.WriteDisposition = bigquery.WriteTruncate

	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		fmt.Println(status.Errors)
		return err
	}

	return nil
}

func TestGetGroupAvgScoresByDate(t *testing.T) {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if err := loadTestDataFromJSON(ctx, client, StudentsTableName, "./testdata/get_group_avg_scores_by_date/students.json"); err != nil {
		t.Fatal(err)
	}
	if err := loadTestDataFromJSON(ctx, client, ScoresTableName, "./testdata/get_group_avg_scores_by_date/scores.json"); err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		haveDate   civil.Date
		wantScores []*groupAvgScore
	}{
		"case1_normal_20190901": {
			haveDate: civil.Date{Year: 2019, Month: 9, Day: 1},
			wantScores: []*groupAvgScore{
				&groupAvgScore{
					Group:    "GroupA",
					AvgScore: float64(40.0),
				},
				&groupAvgScore{
					Group:    "GroupB",
					AvgScore: float64(15.0),
				},
			},
		},
		"case2_no_result_20190902": {
			haveDate:   civil.Date{Year: 2019, Month: 9, Day: 2},
			wantScores: []*groupAvgScore{},
		},
	}

	for k, v := range testCases {
		t.Run(k, func(t *testing.T) {
			gotScores, err := getGroupAvgScoresByDate(ctx, client, v.haveDate)
			if err != nil {
				t.Fatal(err)
			}

			if len(gotScores) != len(v.wantScores) {
				t.Fatal("count of gotScores and wantScores is different.")
			}

			sort.Slice(gotScores, func(i, j int) bool {
				return gotScores[i].Group < gotScores[j].Group
			})
			sort.Slice(v.wantScores, func(i, j int) bool {
				return v.wantScores[i].Group < v.wantScores[j].Group
			})

			for i, want := range v.wantScores {
				if diff := cmp.Diff(*want, *gotScores[i]); diff != "" {
					t.Fatalf("comparing mismatch. want: %v, got: %v, diff: %s", want, gotScores[i], diff)
				}
			}

		})
	}
}
