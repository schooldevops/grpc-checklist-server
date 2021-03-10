package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/schooldevops/go/grpc/checklist/checkpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Checklist Client Up")

	cc, err := grpc.Dial("localhost:10000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	c := checkpb.NewChecklistServiceClient(cc)

	fmt.Println("Create checklist")
	checklist := &checkpb.Checklist{
		Order: 1,
		Channel: "Architecture",
		Code: "AR01",
		Category: "Security",
		Item: "https 로 요청을 처리하고, http는 비활성화 하였는가?",
	}
	createResult, err := c.CreateChecklist(context.Background(), &checkpb.CreateChecklistRequest{Checklist: checklist})

	if err != nil {
		log.Fatalf("Unexped error: %v", err)
	}
	checkListID := createResult.GetResult().GetId()
	fmt.Printf("Checklist Value: %v\n", checkListID);
	
	// 아이디로 조회하기. 
	readChecklistByID, err := c.ReadChecklistByID(context.Background(), &checkpb.ReadChecklistRequest{Id: checkListID})

	if err != nil {
		log.Fatalf("Cannot read checklist: %v", err)
	}
	fmt.Printf("Read Checklist %v\n", readChecklistByID)

	// 쿼리 조건으로 조회하기. 
	stream, err := c.ReadChecklistByQuery(context.Background(), &checkpb.ReadChecklistQueryRequest{
		Query: &checkpb.Checklist{
			Category: "Security",
		},
	})

	if err != nil {
		log.Fatalf("error while caling Checklist: %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break;
		}
		if err != nil {
			log.Fatalf("Error happen %v", err)
		}

		fmt.Printf("Result: %v \n", res.GetResult())
	}

	// 체크 리스트 업데이트 수행하기. 
	modifyChecklist := &checkpb.Checklist{
		Id: checkListID,
		Category: "Cost",
		Order: 2,
	}

	updateResult, updateErr := c.UpdateChecklist(context.Background(), &checkpb.UpdateChecklistRequest{Checklist: modifyChecklist})
	if updateErr != nil {
		fmt.Printf("Error happened while updating: %v \n", updateErr)
	}
	fmt.Printf("Blog was updated: %v\n", updateResult)

	// 체크리스트 삭제하기. 
	deleteRes, deleteErr := c.DeleteCheckist(context.Background(), &checkpb.DeleteChecklistRequest{Id: checkListID})

	if deleteErr != nil {
		fmt.Printf("Error happened while deleting: %v \n", deleteErr)
	}
	fmt.Printf("Checklist was deleted: %v \n", deleteRes)

	// 전체 체크리스트 목록 보기. 

	allStream, err := c.AllCheclkists(context.Background(), &checkpb.ListChecklistRequest{})
	if err != nil {
		log.Fatalf("error while calling ListBlog RPC: %v", err)
	}
	for {
		res, err := allStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(res.GetResult())
	}

	//	벌크 체크리스트 등록하기. 
	checklistArray := []*checkpb.CreateChecklistRequest{
		&checkpb.CreateChecklistRequest{
			Checklist: &checkpb.Checklist{
				Order: 1,
				Channel: "Architecture",
				Code: "AR11",
				Category: "Arch",
				Item: "1. https 로 요청을 처리하고, ---- 1?",
			},
		},
		&checkpb.CreateChecklistRequest{
			Checklist: &checkpb.Checklist{
				Order: 2,
				Channel: "Architecture",
				Code: "AR12",
				Category: "Arch",
				Item: "2. https 로 요청을 처리하고, ---- 2?",
			},
		},
		&checkpb.CreateChecklistRequest{
			Checklist: &checkpb.Checklist{
				Order: 3,
				Channel: "Architecture",
				Code: "AR13",
				Category: "Arch",
				Item: "3. https 로 요청을 처리하고, ---- 3?",
			},
		},
	}

	bulkStream, err := c.CreateBulkChecklist(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}

	for _, bulkReq := range checklistArray {
		fmt.Printf("Send Data: %v\n", bulkReq)
		bulkStream.Send(bulkReq)
		time.Sleep(100 * time.Millisecond)
	}

}