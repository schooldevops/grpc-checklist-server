package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/schooldevops/go/grpc/checklist/checkpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

type checklistItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Order    int32              `bson:"order, omitempty"`
	Channel  string             `bson:"channel, omitempty"`
	Code     string             `bson:"code, omitempty"`
	Category string             `bson:"category, omitempty"`
	Item     string             `bson:"item, omitempty"`
}

// checklistServer is main function
func checklistServer(s *grpc.Server) {

	log.Println("Complete register checklist service.")
	checkpb.RegisterChecklistServiceServer(s, &server{})

}

func (*server) CreateChecklist(ctx context.Context, req *checkpb.CreateChecklistRequest) (*checkpb.CreateChecklistResponse, error) {
	log.Printf("Create Checklist %v\n", req)

	checklist := req.GetChecklist()

	payload := checklistItem{
		Order:    checklist.GetOrder(),
		Channel:  checklist.GetChannel(),
		Code:     checklist.GetCode(),
		Category: checklist.GetCategory(),
		Item:     checklist.GetItem(),
	}

	res, err := collection.InsertOne(context.Background(), payload)
	if err != nil {
		log.Fatalf("Insert Error %v", err)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	objID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to MongoDBID"),
		)
	}

	result := &checkpb.Checklist{
		Id:       objID.Hex(),
		Channel:  checklist.GetChannel(),
		Code:     checklist.GetCode(),
		Category: checklist.GetCategory(),
		Item:     checklist.GetItem(),
	}

	return &checkpb.CreateChecklistResponse{
		Result: result,
	}, nil
}

func (*server) ReadChecklistByID(ctx context.Context, req *checkpb.ReadChecklistRequest) (*checkpb.ReadChecklistResponse, error) {
	checklistID := req.GetId()
	oid, err := primitive.ObjectIDFromHex(checklistID)

	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	checklist := &checklistItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(checklist); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find checklist with specified ID: %v", err),
		)
	}
	// 인텔리안테크
	return &checkpb.ReadChecklistResponse{
		Result: dataToChecklistPb(checklist),
	}, nil
}

func (*server) ReadChecklistByQuery(req *checkpb.ReadChecklistQueryRequest, stream checkpb.ChecklistService_ReadChecklistByQueryServer) error {
	qry := req.GetQuery()

	filter := bson.M{}

	if qry.GetOrder() != 0 {
		filter["order"] = qry.GetOrder()
	}

	if qry.GetChannel() != "" {
		filter["channel"] = qry.GetChannel()
	}
	if qry.GetCode() != "" {
		filter["code"] = qry.GetCode()
	}

	if qry.GetCategory() != "" {
		filter["category"] = qry.GetCategory()
	}

	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}

	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		checklist := &checklistItem{}
		err := cur.Decode(checklist)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)
		}
		stream.Send(&checkpb.ReadChecklistResponse{Result: dataToChecklistPb(checklist)})
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	return nil
}

func (*server) UpdateChecklist(ctx context.Context, req *checkpb.UpdateChecklistRequest) (*checkpb.UpdateChecklistResponse, error) {
	updateValue := req.GetChecklist()

	oid, err := primitive.ObjectIDFromHex(updateValue.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	checklist := &checklistItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(checklist); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}

	if updateValue.GetOrder() != 0 {
		checklist.Order = updateValue.GetOrder()
	}

	if updateValue.GetChannel() != "" {
		checklist.Channel = updateValue.GetChannel()
	}

	if updateValue.GetCode() != "" {
		checklist.Code = updateValue.GetCode()
	}

	if updateValue.GetCategory() != "" {
		checklist.Category = updateValue.GetCategory()
	}

	if updateValue.GetItem() != "" {
		checklist.Item = updateValue.GetItem()
	}

	_, updateErr := collection.ReplaceOne(context.Background(), filter, checklist)

	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object in MongoDB: %v", updateErr),
		)
	}

	return &checkpb.UpdateChecklistResponse{
		Result: dataToChecklistPb(checklist),
	}, nil
}

func (*server) DeleteCheckist(ctx context.Context, req *checkpb.DeleteChecklistRequest) (*checkpb.DeleteChecklistResponse, error) {

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	filter := bson.M{"_id": oid}

	res, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object in MongoDB: %v", err),
		)
	}

	if res.DeletedCount == 0 {
		return &checkpb.DeleteChecklistResponse{Result: false}, nil
	}

	return &checkpb.DeleteChecklistResponse{Result: true}, nil
}

func (*server) AllCheclkists(req *checkpb.ListChecklistRequest, stream checkpb.ChecklistService_AllCheclkistsServer) error {
	cur, err := collection.Find(context.Background(), primitive.D{{}})

	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		checklist := &checklistItem{}
		err := cur.Decode(checklist)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)

		}
		stream.Send(&checkpb.ListChecklistResponse{Result: dataToChecklistPb(checklist)})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	return nil
}

func (*server) Echo(ctx context.Context, in *checkpb.StringMessage) (*checkpb.StringMessage, error) {
	return &checkpb.StringMessage{
		Value: "Hello World",
	}, nil
}

func (*server) CreateBulkChecklist(stream checkpb.ChecklistService_CreateBulkChecklistServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading bulk checklist %v", err)
		}

		checklist := req.GetChecklist()

		payload := checklistItem{
			Order:    checklist.GetOrder(),
			Channel:  checklist.GetChannel(),
			Code:     checklist.GetCode(),
			Category: checklist.GetCategory(),
			Item:     checklist.GetItem(),
		}

		res, err := collection.InsertOne(context.Background(), payload)
		if err != nil {
			log.Fatalf("Insert Error %v", err)
			return err
		}

		objID, ok := res.InsertedID.(primitive.ObjectID)
		if !ok {
			return err
		}

		result := &checkpb.Checklist{
			Id:       objID.Hex(),
			Channel:  checklist.GetChannel(),
			Code:     checklist.GetCode(),
			Category: checklist.GetCategory(),
			Item:     checklist.GetItem(),
		}

		sendErr := stream.Send(&checkpb.CreateChecklistResponse{Result: result})
		if sendErr != nil {
			log.Fatalf("Error while sending data to client: %v", err)
			return err
		}

	}
}

func dataToChecklistPb(data *checklistItem) *checkpb.Checklist {
	return &checkpb.Checklist{
		Id:       data.ID.Hex(),
		Order:    data.Order,
		Channel:  data.Channel,
		Code:     data.Code,
		Category: data.Category,
		Item:     data.Item,
	}
}
