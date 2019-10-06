package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"google.golang.org/api/iterator"
)

var projectID string
var dataset string

func init() {
	projectID = os.Getenv("GCP_PROJECT")
	dataset = os.Getenv("BQ_DATASET")
}

func main() {
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

	date := civil.Date{Year: 2019, Month: 9, Day: 11}
	scores, err := getGroupAvgScoresByDate(ctx, client, date)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range scores {
		fmt.Println(*v)
	}
}

type groupAvgScore struct {
	Group    string
	AvgScore float64
}

func getGroupAvgScoresByDate(ctx context.Context, client *bigquery.Client, date civil.Date) ([]*groupAvgScore, error) {
	q := "SELECT " +
		"st.group,  " +
		"avg(sc.score) as AvgScore  " +
		"FROM `" + projectID + "." + dataset + ".scores` as sc " +
		"INNER JOIN `" + projectID + "." + dataset + ".students` as st ON(st.id = sc.student_id) " +
		"WHERE sc.date = '" + date.String() + "' " +
		"GROUP BY st.group "

	it, err := client.Query(q).Read(ctx)
	if err != nil {
		return nil, err
	}

	var scores []*groupAvgScore
	for {
		var s groupAvgScore
		err := it.Next(&s)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		scores = append(scores, &s)
	}

	return scores, nil
}
