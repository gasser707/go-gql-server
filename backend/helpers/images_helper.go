package helpers

import (
	"fmt"
	"github.com/gasser707/go-gql-server/graphql/model"
	"strings"
)

func ParseLabels(labels []string) string {
	str := ""
	for _, l := range labels {
		str = str + fmt.Sprintf("'%s',", strings.ToLower(l))
	}

	str = str[0 : len(str)-1]
	str = "(" + str + ")"
	return str
}

func ParseFilter(input *model.ImageFilterInput, userID int) string {
	queryStr := []string{}
	filterStart := ""
	filterStr := ""
	filterAdded := false
	if input.Labels != nil && len(input.Labels) > 0 && input.MatchAll != nil && *input.MatchAll {
		filterStr = "select distinct images.id, created_at, url, description, user_id, title, price, forSale, private From labels join images on images.id=labels.image_id where "
		matcher := []string{}
		for _, label := range input.Labels {
			matcher = append(matcher, fmt.Sprintf("image_id in (select image_id from labels where labels.tag='%s')", label))
		}
		filterStr = filterStr + strings.Join(matcher[:], " And ")
		queryStr = append(queryStr, filterStr)
		filterAdded = true

	} else if input.Labels != nil && len(input.Labels) > 0 {
		filterStr = "select images.id, created_at, url, description, user_id, title, price, forSale, private From labels join images on images.id=labels.image_id where "
		filterStr = filterStr + "labels.tag in " + ParseLabels(input.Labels)
		queryStr = append(queryStr, filterStr)
		filterAdded = true

	} else {
		filterStart = "select * from images where "
	}
	if input.UserID != nil && input.Private != nil && fmt.Sprintf("%v", userID) == *input.UserID {
		filterStr = "images.private=" + fmt.Sprintf("%t", *input.Private)
		queryStr = append(queryStr, filterStr)
		filterAdded = true

	} else if input.UserID != nil {
		filterStr = "images.user_id=" + *input.UserID
		if fmt.Sprintf("%v", userID) != *input.UserID {
			filterStr = filterStr + "And images.private=False And images.archived=False"
		}
		queryStr = append(queryStr, filterStr)
		filterAdded = true
	}

	if input.ForSale != nil {
		filterStr = "images.forSale=" + fmt.Sprintf("%t", *input.ForSale)
		queryStr = append(queryStr, filterStr)
		filterAdded = true
	}
	if input.PriceLimit != nil {
		filterStr = "images.price<=" + fmt.Sprintf("%v", input.PriceLimit)
		queryStr = append(queryStr, filterStr)
		filterAdded = true
	}
	if input.DiscountPercentLimit != nil {
		filterStr = "images.discountPercent<=" + fmt.Sprintf("%v", input.DiscountPercentLimit)
		queryStr = append(queryStr, filterStr)
		filterAdded = true
	}
	if input.Title != nil {
		filterStr = "LOWER(images.title) like" + fmt.Sprintf("'%s'", "%"+strings.ToLower(*input.Title)+"%")
		queryStr = append(queryStr, filterStr)
		filterAdded = true
	}
	if input.UserID != nil && fmt.Sprintf("%v", userID) == *input.UserID && input.Archived != nil && *input.Archived == true {
		filterStr = "images.archived=True"
		queryStr = append(queryStr, filterStr)
		filterAdded = true
	}
	if !filterAdded {
		return filterStart + "images.private=False And images.archived=False"
	}

	return filterStart + strings.Join(queryStr[:], " And ")
}

func RemoveDuplicateLabels(newLabels []string, oldLabels []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range oldLabels {
		allKeys[item] = true
	}
	for _, item := range newLabels {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
