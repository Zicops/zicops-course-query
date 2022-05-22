package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

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
		log.Errorf("error adding sub categotries: %v", err)
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
		log.Errorf("error getting course: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetCourseModules(ctx context.Context, courseID *string) ([]*model.Module, error) {
	resp, err := handlers.GetModulesCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting modules: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetModuleByID(ctx context.Context, moduleID *string) (*model.Module, error) {
	resp, err := handlers.GetModuleByID(ctx, moduleID)
	if err != nil {
		log.Errorf("error getting module by id: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetCourseChapters(ctx context.Context, courseID *string) ([]*model.Chapter, error) {
	resp, err := handlers.GetChaptersCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting chapters: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetChapterByID(ctx context.Context, chapterID *string) (*model.Chapter, error) {
	resp, err := handlers.GetChapterByID(ctx, chapterID)
	if err != nil {
		log.Errorf("error getting chapter by id: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetTopics(ctx context.Context, courseID *string) ([]*model.Topic, error) {
	resp, err := handlers.GetTopicsCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topics: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetTopicByID(ctx context.Context, topicID *string) (*model.Topic, error) {
	resp, err := handlers.GetTopicByID(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic by id: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetTopicContent(ctx context.Context, topicID *string) ([]*model.TopicContent, error) {
	resp, err := handlers.GetTopicContent(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic content: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetTopicResources(ctx context.Context, topicID *string) ([]*model.TopicResource, error) {
	resp, err := handlers.GetTopicResources(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic resources: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetTopicQuizes(ctx context.Context, topicID *string) ([]*model.Quiz, error) {
	resp, err := handlers.GetTopicQuizes(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic quizes: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetQuizFiles(ctx context.Context, quizID *string) ([]*model.QuizFile, error) {
	resp, err := handlers.GetQuizFiles(ctx, quizID)
	if err != nil {
		log.Errorf("error getting quiz files: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetMCQQuiz(ctx context.Context, quizID *string) ([]*model.QuizMcq, error) {
	resp, err := handlers.GetMCQQuiz(ctx, quizID)
	if err != nil {
		log.Errorf("error getting mcq quizes: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetDescriptiveQuiz(ctx context.Context, quizID *string) ([]*model.QuizDescriptive, error) {
	resp, err := handlers.GetQuizDes(ctx, quizID)
	if err != nil {
		log.Errorf("error getting descriptive quizes: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetTopicContentByCourseID(ctx context.Context, courseID *string) ([]*model.TopicContent, error) {
	resp, err := handlers.GetTopicContentByCourse(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topic content: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetResourcesByCourseID(ctx context.Context, courseID *string) ([]*model.TopicResource, error) {
	resp, err := handlers.GetCourseResources(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topic resources by course id: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetLatestQuestionBank(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedQuestionBank, error) {
	resp, err := handlers.LatestQuestionBanks(ctx, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("error getting latest question banks: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetQuestionBankQuestions(ctx context.Context, questionBankID *string) ([]*model.QuestionBankQuestion, error) {
	resp, err := handlers.GetQuestionBankQuestions(ctx, questionBankID)
	if err != nil {
		log.Errorf("error getting questions: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetLatestQuestionPapers(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedQuestionPapers, error) {
	resp, err := handlers.LatestQuestionPapers(ctx, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("error getting question papers: %v", err)
		return nil, err
	}
	return resp, nil
}

func (r *queryResolver) GetQuestionPaperSections(ctx context.Context, questionPaperID *string) ([]*model.QuestionPaperSection, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetQPBankMappingByQPId(ctx context.Context, questionPaperID *string) ([]*model.SectionQBMapping, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetQPBankMappingBySectionID(ctx context.Context, sectionID *string) ([]*model.SectionQBMapping, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetSectionFixedQuestions(ctx context.Context, sectionID *string) ([]*model.SectionFixedQuestions, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetOptionsForQuestions(ctx context.Context, questionIds []*string) ([]*model.MapQuestionWithOption, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetExamsByQPId(ctx context.Context, questionPaperID *string) ([]*model.Exam, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetExamSchedule(ctx context.Context, examID *string) (*model.ExamSchedule, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetExamInstruction(ctx context.Context, examID *string) (*model.ExamInstruction, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetExamCohort(ctx context.Context, examID *string) (*model.ExamCohort, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetExamConfiguration(ctx context.Context, examID *string) (*model.ExamConfiguration, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
