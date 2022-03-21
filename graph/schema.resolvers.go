package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-course-query/graph/generated"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/handlers"
)

func (r *queryResolver) AllCategories(ctx context.Context) ([]*string, error) {
	resp, err := handlers.GetCategories(ctx)
	if err != nil {
		log.Errorf("error adding categotries: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) AllSubCategories(ctx context.Context) ([]*string, error) {
	resp, err := handlers.GetSubCategories(ctx)
	if err != nil {
		log.Errorf("error adding categotries: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) LatestCourses(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, status *model.Status) (*model.PaginatedCourse, error) {
	resp, err := handlers.LatestCourses(ctx, publishTime, pageCursor, direction, pageSize, status)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetCourse(ctx context.Context, courseID *string) (*model.Course, error) {
	resp, err := handlers.GetCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetCourseModules(ctx context.Context, courseID *string) ([]*model.Module, error) {
	resp, err := handlers.GetModulesCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetModuleByID(ctx context.Context, moduleID *string) (*model.Module, error) {
	resp, err := handlers.GetModuleByID(ctx, moduleID)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetCourseChapters(ctx context.Context, courseID *string) ([]*model.Chapter, error) {
	resp, err := handlers.GetChaptersCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetChapterByID(ctx context.Context, chapterID *string) (*model.Chapter, error) {
	resp, err := handlers.GetChapterByID(ctx, chapterID)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
type mutationResolver struct{ *Resolver }
