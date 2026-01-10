package tool

import (
	"fmt"
	"strconv"
)

// 写一个函数，输入为当前该月的平均工时，输出为该月剩余工作日的最佳每日工作时间。计算规则为每日工时为9加5分，为10加10分数。要考虑节假日，
// 并且每日工作时间是在上午9点到下午6点之间，午休1.5h。工作为大小周，且考虑法定节假日
func CalculateBestWorkTime(avgMonthlyHours float64, remainingWorkDays int) string {
	// 每日工作时间范围：9:00-12:00，13:30-18:00，共8.5小时
	const workStartHour = 9
	const workEndHour = 18
	const lunchBreakHours = 1.5
	const totalDailyWorkHours = workEndHour - workStartHour - lunchBreakHours // 7.5小时
	if remainingWorkDays <= 0 {
		return "无剩余工作日"
	}
	// 计算每日最佳工作时间
	bestDailyWorkHours := avgMonthlyHours / float64(remainingWorkDays)
	if bestDailyWorkHours > totalDailyWorkHours {
		bestDailyWorkHours = totalDailyWorkHours
	}
	// 将工作时间转换为小时和分钟
	hours := int(bestDailyWorkHours)
	minutes := int((bestDailyWorkHours - float64(hours)) * 60)
	return formatWorkTime(hours, minutes)
}

func formatWorkTime(hours int, minutes int) string {
	startHour := 9
	startMinute := 0
	endHour := startHour + hours
	endMinute := startMinute + minutes
	if endMinute >= 60 {
		endHour += endMinute / 60
		endMinute = endMinute % 60
	}
	return formatTime(startHour, startMinute) + " - " + formatTime(endHour, endMinute)
}

func formatTime(hour int, minute int) string {
	period := "AM"
	if hour >= 12 {
		period = "PM"
	}
	displayHour := hour
	if hour > 12 {
		displayHour = hour - 12
	}
	return strconv.Itoa(displayHour) + ":" + fmt.Sprintf("%02d", minute) + " " + period
}
