package server

// import (
// 	"context"
// 	"log"

// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"

// 	pkgerrors "github.com/coredgeio/compass/pkg/errors"
// 	"github.com/coredgeio/compass/pkg/utils"

// 	api "github.com/coredgeio/workflow-manager/api/service"
// 	// Import your runtime tables if using MongoDB
// 	// "github.com/coredgeio/workflow-manager/pkg/runtime/resource"
// )

// type ServiceApiServer struct {
// 	api.UnimplementedServiceApiServer
// 	// Add your dependencies here
// 	// resourceTbl *resource.ResourceTable
// }

// func NewServiceApiServer() *ServiceApiServer {
// 	// Initialize your tables/dependencies here
// 	// resourceTbl, err := resource.LocateResourceTable()
// 	// if err != nil {
// 	// 	log.Fatalln("ServiceApiServer: failed to locate resource table", err)
// 	// }

// 	return &ServiceApiServer{
// 		// resourceTbl: resourceTbl,
// 	}
// }

// func (s *ServiceApiServer) ListResources(ctx context.Context, req *api.ListResourcesReq) (*api.ListResourcesResp, error) {
// 	// Implement your list logic here
// 	// Example:
// 	// count, err := s.resourceTbl.GetCountInProject(req.Domain, req.Project)
// 	// if err != nil {
// 	// 	log.Println("error fetching resource count in project", err)
// 	// 	return nil, status.Errorf(codes.Internal, "Something went wrong, please try again")
// 	// }

// 	resp := &api.ListResourcesResp{
// 		Count: 0,
// 		Items: []*api.ResourceEntry{},
// 	}

// 	// Add your implementation here

// 	return resp, nil
// }

// func (s *ServiceApiServer) GetResource(ctx context.Context, req *api.GetResourceReq) (*api.GetResourceResp, error) {
// 	// Implement your get logic here
// 	// Example:
// 	// key := &resource.ResourceKey{
// 	// 	Domain:  req.Domain,
// 	// 	Project: req.Project,
// 	// 	Name:    req.Name,
// 	// }
// 	//
// 	// entry, err := s.resourceTbl.Find(key)
// 	// if err != nil {
// 	// 	if pkgerrors.IsNotFound(err) {
// 	// 		return nil, status.Errorf(codes.NotFound, "Entry %q not found", *key)
// 	// 	}
// 	// 	log.Println("Error finding resource", *key, "error", err)
// 	// 	return nil, status.Errorf(codes.Internal, "Something went wrong, please try again")
// 	// }

// 	resp := &api.GetResourceResp{
// 		Name: req.Name,
// 		// Add your response fields here
// 	}

// 	return resp, nil
// }

// func (s *ServiceApiServer) CreateResource(ctx context.Context, req *api.CreateResourceReq) (*api.CreateResourceResp, error) {
// 	// Implement your create logic here
// 	// Example:
// 	// entry := &resource.ResourceEntry{
// 	// 	Key: resource.ResourceKey{
// 	// 		Domain:  req.Domain,
// 	// 		Project: req.Project,
// 	// 		Name:    req.Name,
// 	// 	},
// 	// 	Desc: utils.PString(req.Desc),
// 	// 	Tags: req.Tags,
// 	// }
// 	//
// 	// err := s.resourceTbl.Create(entry)
// 	// if err != nil {
// 	// 	log.Println("Error creating resource", entry.Key, "error", err)
// 	// 	return nil, status.Errorf(codes.Internal, "Something went wrong, please try again")
// 	// }

// 	return &api.CreateResourceResp{}, nil
// }

// func (s *ServiceApiServer) UpdateResource(ctx context.Context, req *api.UpdateResourceReq) (*api.UpdateResourceResp, error) {
// 	// Implement your update logic here
// 	// Example:
// 	// entry := &resource.ResourceEntry{
// 	// 	Key: resource.ResourceKey{
// 	// 		Domain:  req.Domain,
// 	// 		Project: req.Project,
// 	// 		Name:    req.Name,
// 	// 	},
// 	// 	Desc: utils.PString(req.Desc),
// 	// 	Tags: req.Tags,
// 	// }
// 	//
// 	// err := s.resourceTbl.Update(entry)
// 	// if err != nil {
// 	// 	if pkgerrors.IsNotFound(err) {
// 	// 		return nil, status.Errorf(codes.NotFound, "Entry %q not found", entry.Key)
// 	// 	}
// 	// 	log.Println("Error updating resource", entry.Key, "error", err)
// 	// 	return nil, status.Errorf(codes.Internal, "Something went wrong, please try again")
// 	// }

// 	return &api.UpdateResourceResp{}, nil
// }

// func (s *ServiceApiServer) DeleteResource(ctx context.Context, req *api.DeleteResourceReq) (*api.DeleteResourceResp, error) {
// 	// Implement your delete logic here
// 	// Example:
// 	// entry := &resource.ResourceEntry{
// 	// 	Key: resource.ResourceKey{
// 	// 		Domain:  req.Domain,
// 	// 		Project: req.Project,
// 	// 		Name:    req.Name,
// 	// 	},
// 	// 	IsDeleted: true,
// 	// }
// 	//
// 	// err := s.resourceTbl.Update(entry)
// 	// if err != nil {
// 	// 	if pkgerrors.IsNotFound(err) {
// 	// 		return nil, status.Errorf(codes.NotFound, "Entry %q not found", entry.Key)
// 	// 	}
// 	// 	log.Println("Error deleting resource", entry.Key, "error", err)
// 	// 	return nil, status.Errorf(codes.Internal, "Something went wrong, please try again")
// 	// }

// 	return &api.DeleteResourceResp{}, nil
// }
