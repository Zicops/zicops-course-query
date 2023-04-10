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

// AllCatMain is the resolver for the allCatMain field.
func (r *queryResolver) AllCatMain(ctx context.Context, lspIds []*string, searchText *string) ([]*model.CatMain, error) {
	resp, err := handlers.AllCatMain(ctx, lspIds, searchText)
	if err != nil {
		log.Errorf("error getting categories: %v", err)
		return nil, err
	}
	return resp, nil
}

// AllSubCatMain is the resolver for the allSubCatMain field.
func (r *queryResolver) AllSubCatMain(ctx context.Context, lspIds []*string, searchText *string) ([]*model.SubCatMain, error) {
	resp, err := handlers.AllSubCatMain(ctx, lspIds, searchText)
	if err != nil {
		log.Errorf("error getting sub categories: %v", err)
		return nil, err
	}
	return resp, nil
}

// AllSubCatByCatID is the resolver for the allSubCatByCatId field.
func (r *queryResolver) AllSubCatByCatID(ctx context.Context, catID *string) ([]*model.SubCatMain, error) {
	resp, err := handlers.AllSubCatByCatID(ctx, catID)
	if err != nil {
		log.Errorf("error getting sub categories: %v", err)
		return nil, err
	}
	return resp, nil
}

// AllCategories is the resolver for the allCategories field.
func (r *queryResolver) AllCategories(ctx context.Context) ([]*string, error) {
	resp, err := handlers.GetCategories(ctx)
	if err != nil {
		log.Errorf("error adding categories: %v", err)
		return nil, err
	}
	return resp, nil
}

// AllSubCategories is the resolver for the allSubCategories field.
func (r *queryResolver) AllSubCategories(ctx context.Context) ([]*string, error) {
	resp, err := handlers.GetSubCategories(ctx)
	if err != nil {
		log.Errorf("error adding sub categories: %v", err)
		return nil, err
	}
	return resp, nil
}

// AllSubCatsByCat is the resolver for the allSubCatsByCat field.
func (r *queryResolver) AllSubCatsByCat(ctx context.Context, category *string) ([]*string, error) {
	resp, err := handlers.GetSubCategoriesForSub(ctx, category)
	if err != nil {
		log.Errorf("error adding sub categories: %v", err)
		return nil, err
	}
	return resp, nil
}

// LatestCourses is the resolver for the latestCourses field.
func (r *queryResolver) LatestCourses(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, status *model.Status, filters *model.CoursesFilters) (*model.PaginatedCourse, error) {
	resp, err := handlers.LatestCourses(ctx, publishTime, pageCursor, direction, pageSize, status, filters)
	if err != nil {
		log.Errorf("error getting latest courses: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCourse is the resolver for the getCourse field.
func (r *queryResolver) GetCourse(ctx context.Context, courseID []*string) ([]*model.Course, error) {
	resp, err := handlers.GetCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting course: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCourseModules is the resolver for the getCourseModules field.
func (r *queryResolver) GetCourseModules(ctx context.Context, courseID *string) ([]*model.Module, error) {
	resp, err := handlers.GetModulesCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting modules: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetModuleByID is the resolver for the getModuleById field.
func (r *queryResolver) GetModuleByID(ctx context.Context, moduleID *string) (*model.Module, error) {
	resp, err := handlers.GetModuleByID(ctx, moduleID)
	if err != nil {
		log.Errorf("error getting module by id: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCourseChapters is the resolver for the getCourseChapters field.
func (r *queryResolver) GetCourseChapters(ctx context.Context, courseID *string) ([]*model.Chapter, error) {
	resp, err := handlers.GetChaptersCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting chapters: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetChapterByID is the resolver for the getChapterById field.
func (r *queryResolver) GetChapterByID(ctx context.Context, chapterID *string) (*model.Chapter, error) {
	resp, err := handlers.GetChapterByID(ctx, chapterID)
	if err != nil {
		log.Errorf("error getting chapter by id: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopics is the resolver for the getTopics field.
func (r *queryResolver) GetTopics(ctx context.Context, courseID *string) ([]*model.Topic, error) {
	resp, err := handlers.GetTopicsCourseByID(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topics: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicByID is the resolver for the getTopicById field.
func (r *queryResolver) GetTopicByID(ctx context.Context, topicID *string) (*model.Topic, error) {
	resp, err := handlers.GetTopicByID(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic by id: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicContent is the resolver for the getTopicContent field.
func (r *queryResolver) GetTopicContent(ctx context.Context, topicID *string) ([]*model.TopicContent, error) {
	resp, err := handlers.GetTopicContent(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic content: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicExams is the resolver for the getTopicExams field.
func (r *queryResolver) GetTopicExams(ctx context.Context, topicID *string) ([]*model.TopicExam, error) {
	resp, err := handlers.GetTopicExams(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic content: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicResources is the resolver for the getTopicResources field.
func (r *queryResolver) GetTopicResources(ctx context.Context, topicID *string) ([]*model.TopicResource, error) {
	resp, err := handlers.GetTopicResources(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic resources: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicQuizes is the resolver for the getTopicQuizes field.
func (r *queryResolver) GetTopicQuizes(ctx context.Context, topicID *string) ([]*model.Quiz, error) {
	resp, err := handlers.GetTopicQuizes(ctx, topicID)
	if err != nil {
		log.Errorf("error getting topic quizes: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQuizFiles is the resolver for the getQuizFiles field.
func (r *queryResolver) GetQuizFiles(ctx context.Context, quizID *string) ([]*model.QuizFile, error) {
	resp, err := handlers.GetQuizFiles(ctx, quizID)
	if err != nil {
		log.Errorf("error getting quiz files: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetMCQQuiz is the resolver for the getMCQQuiz field.
func (r *queryResolver) GetMCQQuiz(ctx context.Context, quizID *string) ([]*model.QuizMcq, error) {
	resp, err := handlers.GetMCQQuiz(ctx, quizID)
	if err != nil {
		log.Errorf("error getting mcq quizes: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetDescriptiveQuiz is the resolver for the getDescriptiveQuiz field.
func (r *queryResolver) GetDescriptiveQuiz(ctx context.Context, quizID *string) ([]*model.QuizDescriptive, error) {
	resp, err := handlers.GetQuizDes(ctx, quizID)
	if err != nil {
		log.Errorf("error getting descriptive quizes: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicContentByCourseID is the resolver for the getTopicContentByCourseId field.
func (r *queryResolver) GetTopicContentByCourseID(ctx context.Context, courseID *string) ([]*model.TopicContent, error) {
	resp, err := handlers.GetTopicContentByCourse(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topic content: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicContentByModuleID is the resolver for the getTopicContentByModuleId field.
func (r *queryResolver) GetTopicContentByModuleID(ctx context.Context, moduleID *string) ([]*model.TopicContent, error) {
	resp, err := handlers.GetTopicContentByModule(ctx, moduleID)
	if err != nil {
		log.Errorf("error getting topic content: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicExamsByCourseID is the resolver for the getTopicExamsByCourseId field.
func (r *queryResolver) GetTopicExamsByCourseID(ctx context.Context, courseID *string) ([]*model.TopicExam, error) {
	resp, err := handlers.GetTopicExamsByCourse(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topic exams: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetResourcesByCourseID is the resolver for the getResourcesByCourseId field.
func (r *queryResolver) GetResourcesByCourseID(ctx context.Context, courseID *string) ([]*model.TopicResource, error) {
	resp, err := handlers.GetCourseResources(ctx, courseID)
	if err != nil {
		log.Errorf("error getting topic resources by course id: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetLatestQuestionBank is the resolver for the getLatestQuestionBank field.
func (r *queryResolver) GetLatestQuestionBank(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string, lspID *string) (*model.PaginatedQuestionBank, error) {
	resp, err := handlers.LatestQuestionBanks(ctx, publishTime, pageCursor, direction, pageSize, searchText, lspID)
	if err != nil {
		log.Errorf("error getting latest question banks: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQBMeta is the resolver for the getQBMeta field.
func (r *queryResolver) GetQBMeta(ctx context.Context, qbIds []*string) ([]*model.QuestionBank, error) {
	resp, err := handlers.GetQBMeta(ctx, qbIds)
	if err != nil {
		log.Errorf("error getting latest question banks: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQuestionBankQuestions is the resolver for the getQuestionBankQuestions field.
func (r *queryResolver) GetQuestionBankQuestions(ctx context.Context, questionBankID *string, filters *model.QBFilters) ([]*model.QuestionBankQuestion, error) {
	resp, err := handlers.GetQuestionBankQuestions(ctx, questionBankID, filters)
	if err != nil {
		log.Errorf("error getting questions: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetLatestQuestionPapers is the resolver for the getLatestQuestionPapers field.
func (r *queryResolver) GetLatestQuestionPapers(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedQuestionPapers, error) {
	resp, err := handlers.LatestQuestionPapers(ctx, publishTime, pageCursor, direction, pageSize, searchText)
	if err != nil {
		log.Errorf("error getting question papers: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQPMeta is the resolver for the getQPMeta field.
func (r *queryResolver) GetQPMeta(ctx context.Context, questionPapersIds []*string) ([]*model.QuestionPaper, error) {
	resp, err := handlers.GetQPMeta(ctx, questionPapersIds)
	if err != nil {
		log.Errorf("error getting question papers: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetLatestExams is the resolver for the getLatestExams field.
func (r *queryResolver) GetLatestExams(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedExams, error) {
	resp, err := handlers.GetLatestExams(ctx, publishTime, pageCursor, direction, pageSize, searchText)
	if err != nil {
		log.Errorf("error getting exams: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQuestionPaperSections is the resolver for the getQuestionPaperSections field.
func (r *queryResolver) GetQuestionPaperSections(ctx context.Context, questionPaperID *string) ([]*model.QuestionPaperSection, error) {
	resp, err := handlers.GetQuestionBankSections(ctx, questionPaperID)
	if err != nil {
		log.Errorf("error getting question papers sections: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQPBankMappingByQBId is the resolver for the getQPBankMappingByQBId field.
func (r *queryResolver) GetQPBankMappingByQBId(ctx context.Context, questionBankID *string) ([]*model.SectionQBMapping, error) {
	resp, err := handlers.GetQPBankMappingByQPId(ctx, questionBankID)
	if err != nil {
		log.Errorf("error getting question papers sections map: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQPBankMappingBySectionID is the resolver for the getQPBankMappingBySectionId field.
func (r *queryResolver) GetQPBankMappingBySectionID(ctx context.Context, sectionID *string) ([]*model.SectionQBMapping, error) {
	resp, err := handlers.GetQPBankMappingBySectionID(ctx, sectionID)
	if err != nil {
		log.Errorf("error getting question papers sections map: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetSectionFixedQuestions is the resolver for the getSectionFixedQuestions field.
func (r *queryResolver) GetSectionFixedQuestions(ctx context.Context, sectionID *string) ([]*model.SectionFixedQuestions, error) {
	resp, err := handlers.GetSectionFixedQuestions(ctx, sectionID)
	if err != nil {
		log.Errorf("error getting question papers sections questions: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetOptionsForQuestions is the resolver for the getOptionsForQuestions field.
func (r *queryResolver) GetOptionsForQuestions(ctx context.Context, questionIds []*string) ([]*model.MapQuestionWithOption, error) {
	resp, err := handlers.GetOptionsForQuestions(ctx, questionIds)
	if err != nil {
		log.Errorf("error getting question options: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamsMeta is the resolver for the getExamsMeta field.
func (r *queryResolver) GetExamsMeta(ctx context.Context, examIds []*string) ([]*model.Exam, error) {
	resp, err := handlers.GetExamsMeta(ctx, examIds)
	if err != nil {
		log.Errorf("error getting exams: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamsByQPId is the resolver for the getExamsByQPId field.
func (r *queryResolver) GetExamsByQPId(ctx context.Context, questionPaperID *string) ([]*model.Exam, error) {
	resp, err := handlers.GetExamsByQPId(ctx, questionPaperID)
	if err != nil {
		log.Errorf("error getting exams: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamSchedule is the resolver for the getExamSchedule field.
func (r *queryResolver) GetExamSchedule(ctx context.Context, examID *string) ([]*model.ExamSchedule, error) {
	resp, err := handlers.GetExamSchedule(ctx, examID)
	if err != nil {
		log.Errorf("error getting exam schedule: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamInstruction is the resolver for the getExamInstruction field.
func (r *queryResolver) GetExamInstruction(ctx context.Context, examID *string) ([]*model.ExamInstruction, error) {
	resp, err := handlers.GetExamInstruction(ctx, examID)
	if err != nil {
		log.Errorf("error getting exam instructions: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamCohort is the resolver for the getExamCohort field.
func (r *queryResolver) GetExamCohort(ctx context.Context, examID *string) ([]*model.ExamCohort, error) {
	resp, err := handlers.GetExamCohort(ctx, examID)
	if err != nil {
		log.Errorf("error getting exam cohort: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamConfiguration is the resolver for the getExamConfiguration field.
func (r *queryResolver) GetExamConfiguration(ctx context.Context, examID *string) ([]*model.ExamConfiguration, error) {
	resp, err := handlers.GetExamConfiguration(ctx, examID)
	if err != nil {
		log.Errorf("error getting exam configuration: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetQuestionsByID is the resolver for the getQuestionsById field.
func (r *queryResolver) GetQuestionsByID(ctx context.Context, questionIds []*string) ([]*model.QuestionBankQuestion, error) {
	resp, err := handlers.GetQuestionsByID(ctx, questionIds)
	if err != nil {
		log.Errorf("error getting questions for ids: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCohortCourseMaps is the resolver for the getCohortCourseMaps field.
func (r *queryResolver) GetCohortCourseMaps(ctx context.Context, cohortID *string) ([]*model.CourseCohort, error) {
	resp, err := handlers.GetCohortCourseMaps(ctx, cohortID)
	if err != nil {
		log.Errorf("error getting cohorts for id: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCourseDiscussion is the resolver for the getCourseDiscussion field.
func (r *queryResolver) GetCourseDiscussion(ctx context.Context, courseID string, discussionID *string) ([]*model.Discussion, error) {
	resp, err := handlers.GetCourseDiscussion(ctx, courseID, discussionID)
	if err != nil {
		log.Errorf("error getting course discussion %v", err)
	}
	return resp, err
}

// GetBasicCourseStats is the resolver for the getBasicCourseStats field.
func (r *queryResolver) GetBasicCourseStats(ctx context.Context, input *model.BasicCourseStatsInput) (*model.BasicCourseStats, error) {
	resp, err := handlers.GetBasicCourseStats(ctx, input)
	if err != nil {
		log.Errorf("error getting basic course stats %v", err)
	}
	return resp, err
}

// GetTopicsByCourseIds is the resolver for the getTopicsByCourseIds field.
func (r *queryResolver) GetTopicsByCourseIds(ctx context.Context, courseIds []*string, typeArg *string) ([]*model.Topic, error) {
	resp, err := handlers.GetTopicsByCourseIds(ctx, courseIds, typeArg)
	if err != nil {
		log.Errorf("error getting topics: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetTopicExamsByCourseIds is the resolver for the getTopicExamsByCourseIds field.
func (r *queryResolver) GetTopicExamsByCourseIds(ctx context.Context, courseIds []*string) ([]*model.TopicExam, error) {
	resp, err := handlers.GetTopicExamsByCourseIds(ctx, courseIds)
	if err != nil {
		log.Errorf("error getting topic exams: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamInstructionByExamID is the resolver for the getExamInstructionByExamId field.
func (r *queryResolver) GetExamInstructionByExamID(ctx context.Context, examIds []*string) ([]*model.ExamInstruction, error) {
	resp, err := handlers.GetExamInstructionByExamID(ctx, examIds)
	if err != nil {
		log.Errorf("error getting exam instructions: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetExamScheduleByExamID is the resolver for the getExamScheduleByExamId field.
func (r *queryResolver) GetExamScheduleByExamID(ctx context.Context, examIds []*string) ([]*model.ExamSchedule, error) {
	resp, err := handlers.GetExamScheduleByExamID(ctx, examIds)
	if err != nil {
		log.Errorf("error getting exam schedule: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCourseCountStats is the resolver for the getCourseCountStats field.
func (r *queryResolver) GetCourseCountStats(ctx context.Context, lspID *string, status string, typeArg string) (*model.CourseCountStats, error) {
	resp, err := handlers.GetCourseCountStats(ctx, lspID, status, typeArg)
	if err != nil {
		log.Errorf("error getting course stats: %v", err)
		return nil, err
	}
	return resp, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
