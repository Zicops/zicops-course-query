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
		for _, l := range v.Likes {
			tmp := &l
			likesArray = append(likesArray, tmp)
		}
		var dislikesArray []*string
		for _, d := range v.Dislike {
			tmp := &d
			dislikesArray = append(dislikesArray, tmp)
		}
		t := int(v.Time)
		//here convert time to string in format of minutes:seconds and return
		seconds := t % 60
		minutes := t / 60
		hours := minutes / 60
		var videoTime string
		if hours > 0 {
			videoTime = fmt.Sprintf("%d:%d:%d", hours, minutes, seconds)
		} else {
			videoTime = fmt.Sprintf("%d:%d", minutes, seconds)
		}
		ca := int(v.CreatedAt)
		ua := int(v.UpdatedAt)
		temp := model.Discussion{
			DiscussionID:   &v.DiscussionId,
			CourseID:       &v.CourseId,
			ReplyID:        &v.ReplyId,
			UserID:         &v.UserId,
			Time:           &videoTime,
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

/*
editable
Content,time, likes, dislikes, isanonymous, ispinned, isannouncement, status
*/
