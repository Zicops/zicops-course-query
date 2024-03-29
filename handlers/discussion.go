package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/contracts/coursez"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
)

func GetCourseDiscussion(ctx context.Context, courseID string, discussionID *string) ([]*model.Discussion, error) {

	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	var qryStr string
	// if there is no reply_id in table, only take them first, like reply id == ""
	qryStr = fmt.Sprintf(`SELECT * from coursez.discussion where course_id='%s' `, courseID)
	if discussionID != nil {
		qryStr = fmt.Sprintf(`SELECT * from coursez.discussion where course_id='%s' and reply_id='%s' ALLOW FILTERING`, courseID, *discussionID)
	} else {
		qryStr = fmt.Sprintf(`SELECT * from coursez.discussion where course_id='%s' and reply_id='' ALLOW FILTERING`, courseID)
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

	result := make([]*model.Discussion, len(data))

	for k, vv := range data {

		v := vv
		var likesArray []*string
		for _, ll := range v.Likes {
			l := ll
			tmp := &l
			likesArray = append(likesArray, tmp)
		}
		var dislikesArray []*string
		for _, dd := range v.Dislike {
			d := dd
			tmp := &d
			dislikesArray = append(dislikesArray, tmp)
		}

		t := int(v.Time)
		seconds := t % 60
		minutes := t / 60
		hours := 0
		if minutes >= 60 {
			minutes = minutes % 60
			hours = minutes / 60
		}
		var video_time string
		if hours > 0 {
			video_time = fmt.Sprintf("%d:%d:%d", hours, minutes, seconds)
		} else {
			video_time = fmt.Sprintf("%d:%d", minutes, seconds)
		}
		//here convert time to string in format of minutes:seconds and return
		ca := int(v.CreatedAt)
		ua := int(v.UpdatedAt)
		temp := model.Discussion{
			DiscussionID:   &v.DiscussionId,
			CourseID:       &v.CourseId,
			ReplyID:        &v.ReplyId,
			UserID:         &v.UserId,
			Time:           &video_time,
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
		result[k] = &temp
	}
	return result, nil
}
