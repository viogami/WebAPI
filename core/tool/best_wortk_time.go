package tool

import (
    "fmt"
    "strconv"
    "time"
)

// HolidayConfig 定义节假日配置
type HolidayConfig struct {
    Holidays []string // 格式 "2006-01-02"
    Workdays []string // 补班日期，格式 "2006-01-02"
}

// CalculateBestWorkTime 计算每日最佳工作时间
// isBigWeekStart: true 表示当前周是大周（单休），false 表示小周（双休），后续周交替
func CalculateBestWorkTime(avgMonthlyHours float64, startDate time.Time, endDate time.Time, config HolidayConfig, isBigWeekStart bool) string {
    // 每日工作时间范围：9:00-12:00，13:30-18:00，共7.5小时
    const totalDailyWorkHours = 7.5

    remainingWorkDays := 0
    current := startDate
    
    // 处理节假日Map以便快速查找
    holidayMap := make(map[string]bool)
    for _, h := range config.Holidays {
        holidayMap[h] = true
    }
    workdayMap := make(map[string]bool)
    for _, w := range config.Workdays {
        workdayMap[w] = true
    }

    // 辅助函数：判断是否为大周 (每隔7天切换一次状态)
    // 这里假设 startDate 所在的周即为起始周状态
    // 计算两个日期相差的周数来确定当前是大周还是小周
    _, startWeek := startDate.ISOWeek()

    for !current.After(endDate) {
        dateStr := current.Format("2006-01-02")
        isWorkDay := false

        // 1. 如果是法定调休补班，即使是周末也是工作日
        if workdayMap[dateStr] {
            isWorkDay = true
        } else if holidayMap[dateStr] {
            // 2. 如果是法定节假日，即使是周一到周五也是休息
            isWorkDay = false
        } else {
            // 3. 常规逻辑
            weekday := current.Weekday()
            _, currentWeek := current.ISOWeek()
            weekDiff := currentWeek - startWeek
            
            // 判断当前周是大周还是小周
            isCurrentBigWeek := isBigWeekStart
            if weekDiff%2 != 0 {
                isCurrentBigWeek = !isBigWeekStart
            }

            if weekday >= time.Monday && weekday <= time.Friday {
                isWorkDay = true
            } else if weekday == time.Saturday {
                // 大周周六上班，小周周六休息
                if isCurrentBigWeek {
                    isWorkDay = true
                }
            }
            // 周日通常休息
        }

        if isWorkDay {
            remainingWorkDays++
        }
        current = current.AddDate(0, 0, 1)
    }

    if remainingWorkDays <= 0 {
        return "无剩余有效工作日"
    }

    // 计算每日最佳工作时间
    bestDailyWorkHours := avgMonthlyHours / float64(remainingWorkDays)
    if bestDailyWorkHours > totalDailyWorkHours {
        // 如果平均下来超过最大工时，说明无法完成，或者需要加班，这里简单返回最大工时
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
    endHour := startHour + hours // ...existing code...
    endMinute := startMinute + minutes
    if endMinute >= 60 {
        endHour += endMinute / 60
        endMinute = endMinute % 60
    }
    
    // 处理午休时间 12:00 - 13:30 (1.5小时)
    // 如果结束时间超过12:00，需要顺延1.5小时
    if endHour > 12 || (endHour == 12 && endMinute > 0) {
        endMinute += 30
        if endMinute >= 60 {
            endHour += 1 + endMinute/60
            endMinute = endMinute % 60
        } else {
            endHour += 1
        }
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
