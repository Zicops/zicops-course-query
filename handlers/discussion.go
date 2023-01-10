package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetCourseDiscussion(ctx context.Context, courseID string, discussionID *string) ([]*model.Discussion, error) {

	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * from coursez.discussion where course_id='%s' `, courseID)
	if discussionID == nil && *discussionID == "" {
		qryStr = qryStr + `ALLOW FILTERING`
	} else {
		qryStr = qryStr + fmt.Sprintf(`and discussion_id='%s' ALLOW FILTERING`, *discussionID)
	}
	getDiscussions := func() (modules []coursez.Discussion, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return modules, iter.Select(&modules)
	}

	data, err := getDiscussions()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var result []*model.Discussion

	for _, v := range data {

		var likesArray []*string
		for _, l := range v.Likes {
			likesArray = append(likesArray, &l)
		}
		var dislikesArray []*string
		for _, d := range v.Dislike {
			dislikesArray = append(dislikesArray, &d)
		}
		t := int(v.Time)
		ca := int(v.CreatedAt)
		ua := int(v.UpdatedAt)
		temp := &model.Discussion{
			DiscussionID:   &v.DiscussionId,
			CourseID:       &v.CourseId,
			ReplyID:        &v.ReplyId,
			UserID:         &v.UserId,
			Time:           &t,
			Content:        &v.Content,
			Module:         &v.Module,
			Chapter:        &v.Chapter,
			Topic:          &v.Topic,
			Likes:          likesArray,
			Dislike:        dislikesArray,
			IsAnonymous:    &v.IsAnonymous,
			IsPinned:       &v.IsPinned,
			IsAnnouncement: &v.IsAnnouncement,
			ReplyCount:     &v.ReplyCount,
			CreatedBy:      &v.CreatedBy,
			CreatedAt:      &ca,
			UpdatedBy:      &v.UpdatedBy,
			UpdatedAt:      &ua,
			Status:         &v.Status,
		}
		result = append(result, temp)
	}
	return result, nil
}
