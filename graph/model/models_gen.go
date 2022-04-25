// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type Chapter struct {
	ID          *string `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ModuleID    *string `json:"moduleId"`
	CourseID    *string `json:"courseId"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
	Sequence    *int    `json:"sequence"`
}

type Course struct {
	ID                 *string          `json:"id"`
	Name               *string          `json:"name"`
	Description        *string          `json:"description"`
	Summary            *string          `json:"summary"`
	Instructor         *string          `json:"instructor"`
	Image              *string          `json:"image"`
	PreviewVideo       *string          `json:"previewVideo"`
	TileImage          *string          `json:"tileImage"`
	Owner              *string          `json:"owner"`
	Duration           *int             `json:"duration"`
	ExpertiseLevel     *string          `json:"expertise_level"`
	Language           []*string        `json:"language"`
	Benefits           []*string        `json:"benefits"`
	Outcomes           []*string        `json:"outcomes"`
	CreatedAt          *string          `json:"created_at"`
	UpdatedAt          *string          `json:"updated_at"`
	Type               *string          `json:"type"`
	Prequisites        []*string        `json:"prequisites"`
	GoodFor            []*string        `json:"goodFor"`
	MustFor            []*string        `json:"mustFor"`
	RelatedSkills      []*string        `json:"related_skills"`
	PublishDate        *string          `json:"publish_date"`
	ExpiryDate         *string          `json:"expiry_date"`
	ExpectedCompletion *string          `json:"expected_completion"`
	QaRequired         *bool            `json:"qa_required"`
	Approvers          []*string        `json:"approvers"`
	CreatedBy          *string          `json:"created_by"`
	UpdatedBy          *string          `json:"updated_by"`
	Status             *Status          `json:"status"`
	IsDisplay          *bool            `json:"is_display"`
	Category           *string          `json:"category"`
	SubCategory        *string          `json:"sub_category"`
	SubCategories      []*SubCategories `json:"sub_categories"`
}

type Module struct {
	ID          *string `json:"id"`
	Name        *string `json:"name"`
	IsChapter   *bool   `json:"isChapter"`
	Description *string `json:"description"`
	CourseID    *string `json:"courseId"`
	Owner       *string `json:"owner"`
	Duration    *int    `json:"duration"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
	Level       *string `json:"level"`
	Sequence    *int    `json:"sequence"`
	SetGlobal   *bool   `json:"setGlobal"`
}

type PaginatedCourse struct {
	Courses    []*Course `json:"courses"`
	PageCursor *string   `json:"pageCursor"`
	Direction  *string   `json:"direction"`
	PageSize   *int      `json:"pageSize"`
}

type Quiz struct {
	ID          *string `json:"id"`
	Name        *string `json:"name"`
	Category    *string `json:"category"`
	Type        *string `json:"type"`
	IsMandatory *bool   `json:"isMandatory"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
	TopicID     *string `json:"topicId"`
	Sequence    *int    `json:"sequence"`
	StartTime   *int    `json:"startTime"`
}

type QuizDescriptive struct {
	QuizID        *string `json:"quizId"`
	Question      *string `json:"question"`
	CorrectAnswer *string `json:"correctAnswer"`
	Explanation   *string `json:"explanation"`
}

type QuizFile struct {
	QuizID  *string `json:"quizId"`
	Type    *string `json:"type"`
	Name    *string `json:"name"`
	FileURL *string `json:"fileUrl"`
}

type QuizMcq struct {
	QuizID        *string   `json:"quizId"`
	Question      *string   `json:"question"`
	Options       []*string `json:"options"`
	CorrectOption *string   `json:"correctOption"`
	Explanation   *string   `json:"explanation"`
}

type Topic struct {
	ID          *string `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	ModuleID    *string `json:"moduleId"`
	ChapterID   *string `json:"chapterId"`
	CourseID    *string `json:"courseId"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
	Sequence    *int    `json:"sequence"`
	CreatedBy   *string `json:"created_by"`
	UpdatedBy   *string `json:"updated_by"`
	Image       *string `json:"image"`
}

type TopicContent struct {
	ID                *string `json:"id"`
	Language          *string `json:"language"`
	TopicID           *string `json:"topicId"`
	CourseID          *string `json:"courseId"`
	StartTime         *int    `json:"startTime"`
	Duration          *int    `json:"duration"`
	SkipIntroDuration *int    `json:"skipIntroDuration"`
	NextShowTime      *int    `json:"nextShowTime"`
	FromEndTime       *int    `json:"fromEndTime"`
	CreatedAt         *string `json:"created_at"`
	UpdatedAt         *string `json:"updated_at"`
	Type              *string `json:"type"`
	ContentURL        *string `json:"contentUrl"`
	SubtitleURL       *string `json:"subtitleUrl"`
}

type TopicResource struct {
	ID        *string `json:"id"`
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	TopicID   *string `json:"topicId"`
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
	CreatedBy *string `json:"created_by"`
	UpdatedBy *string `json:"updated_by"`
	URL       *string `json:"url"`
}

type SubCategories struct {
	Name *string `json:"name"`
	Rank *int    `json:"rank"`
}

type Status string

const (
	StatusSaved           Status = "SAVED"
	StatusApprovalPending Status = "APPROVAL_PENDING"
	StatusOnHold          Status = "ON_HOLD"
	StatusApproved        Status = "APPROVED"
	StatusPublished       Status = "PUBLISHED"
	StatusRejected        Status = "REJECTED"
)

var AllStatus = []Status{
	StatusSaved,
	StatusApprovalPending,
	StatusOnHold,
	StatusApproved,
	StatusPublished,
	StatusRejected,
}

func (e Status) IsValid() bool {
	switch e {
	case StatusSaved, StatusApprovalPending, StatusOnHold, StatusApproved, StatusPublished, StatusRejected:
		return true
	}
	return false
}

func (e Status) String() string {
	return string(e)
}

func (e *Status) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Status(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Status", str)
	}
	return nil
}

func (e Status) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
