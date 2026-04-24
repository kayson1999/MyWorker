package utils

import (
	"fmt"
	"strings"
)

// DailyTitle 单日工作风格标签计算结果
type DailyTitle struct {
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	ClockIn  string `json:"clock_in"`
	ClockOut string `json:"clock_out"`
	Duration string `json:"duration"`
	Score    int    `json:"score"`
}

// CalculateDailyTitle 计算单日工作风格标签
func CalculateDailyTitle(clockIn, clockOut string, duration float64) DailyTitle {
	if clockIn == "" || clockOut == "" {
		return DailyTitle{
			Title:    "未完成打卡",
			SubTitle: "请完成上下班打卡",
			Score:    0,
		}
	}

	// 解析时间
	inHour, inMin := parseTime(clockIn)
	outHour, outMin := parseTime(clockOut)

	// 计算时间特征评分
	inScore := getClockInScore(inHour, inMin)
	outScore := getClockOutScore(outHour, outMin)
	durationScore := getDurationScore(duration)

	// 场景化组合匹配
	title, subTitle := matchScenario(inHour, inMin, outHour, outMin, duration)

	return DailyTitle{
		Title:    title,
		SubTitle: subTitle,
		ClockIn:  fmt.Sprintf("%02d:%02d", inHour, inMin),
		ClockOut: fmt.Sprintf("%02d:%02d", outHour, outMin),
		Duration: fmt.Sprintf("%.1fh", duration),
		Score:    (inScore + outScore + durationScore) / 3,
	}
}

// CalculatePeriodTitle 计算周期工作风格标签（周/月/年）
func CalculatePeriodTitle(dailyTitles []DailyTitle) string {
	if len(dailyTitles) == 0 {
		return "新人驾到 🎉"
	}

	// 统计各称号出现频率
	titleCount := make(map[string]int)
	totalScore := 0

	for _, dt := range dailyTitles {
		titleCount[dt.Title]++
		totalScore += dt.Score
	}

	// 找出出现次数最多的称号
	var mostCommonTitle string
	maxCount := 0
	for title, count := range titleCount {
		if count > maxCount {
			maxCount = count
			mostCommonTitle = title
		}
	}

	// 计算平均分决定最终称号前缀
	avgScore := totalScore / len(dailyTitles)

	// 提取称号的纯文本部分（去掉emoji）
	baseName := stripEmoji(mostCommonTitle)

	switch {
	case avgScore >= 90:
		return "终极" + baseName + " 🏆"
	case avgScore >= 80:
		return "资深" + baseName + " ⭐"
	case avgScore >= 70:
		return mostCommonTitle // 保留原始称号（含emoji）
	case avgScore >= 60:
		return baseName
	default:
		return "萌新" + baseName + " 🌱"
	}
}

// stripEmoji 去除称号中的emoji后缀，保留纯文本
func stripEmoji(title string) string {
	// 称号格式为 "文字 emoji"，按最后一个空格分割
	idx := strings.LastIndex(title, " ")
	if idx > 0 {
		return strings.TrimSpace(title[:idx])
	}
	return title
}

// parseTime 解析时间字符串 "HH:MM" 为小时和分钟
func parseTime(timeStr string) (int, int) {
	var hour, min int
	fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
	return hour, min
}

// getClockInScore 上班时间评分（越早分越高）
func getClockInScore(hour, min int) int {
	totalMin := hour*60 + min
	switch {
	case totalMin < 8*60:
		return 95
	case totalMin < 8*60+30:
		return 85
	case totalMin < 9*60:
		return 75
	case totalMin < 9*60+30:
		return 65
	default:
		return 55
	}
}

// getClockOutScore 下班时间评分（越晚分越高）
func getClockOutScore(hour, min int) int {
	totalMin := hour*60 + min
	switch {
	case totalMin >= 22*60:
		return 95
	case totalMin >= 21*60:
		return 85
	case totalMin >= 20*60:
		return 75
	case totalMin >= 18*60+30:
		return 65
	default:
		return 55
	}
}

// getDurationScore 工时评分（越长分越高）
func getDurationScore(duration float64) int {
	switch {
	case duration >= 12:
		return 95
	case duration >= 10:
		return 85
	case duration >= 8:
		return 75
	case duration >= 6:
		return 65
	default:
		return 45
	}
}

// matchScenario 场景化匹配称号（根据上下班时间和工时组合判断）
func matchScenario(inHour, inMin, outHour, outMin int, duration float64) (string, string) {
	inTime := inHour*60 + inMin
	outTime := outHour*60 + outMin

	// 超级肝帝：早到(<=08:00) + 晚走(>=22:00) + 超长工时(>=12h)
	if inTime <= 8*60 && outTime >= 22*60 && duration >= 12 {
		return "超级肝帝 🏆", "从早肝到晚的终极卷王"
	}

	// 深夜战士：晚到(>=09:30) + 晚走(>=22:00) + 超长工时(>=12h)
	if inTime >= 9*60+30 && outTime >= 22*60 && duration >= 12 {
		return "深夜战士 🦇", "夜深人静还在肝"
	}

	// 卷王之王：早到(<=08:30) + 晚走(>=21:00) + 长工时(>=10h)
	if inTime <= 8*60+30 && outTime >= 21*60 && duration >= 10 {
		return "卷王之王 👑", "全方位卷的王者"
	}

	// 高效战士：早到(<=08:30) + 正常走(<=20:00) + 标准工时(8-10h)
	if inTime <= 8*60+30 && outTime <= 20*60 && duration >= 8 && duration <= 10 {
		return "高效战士 ⚔️", "早来早走效率高"
	}

	// 夜猫子：晚到(>=09:30) + 晚走(>=21:00) + 标准工时(8-10h)
	if inTime >= 9*60+30 && outTime >= 21*60 && duration >= 8 && duration <= 10 {
		return "夜猫子 🦉", "晚来晚走精神好"
	}

	// 标准打工人：正常到(08:30-09:30) + 正常走(18:30-20:00) + 标准工时(8-10h)
	if inTime >= 8*60+30 && inTime <= 9*60+30 &&
		outTime >= 18*60+30 && outTime <= 20*60 &&
		duration >= 8 && duration <= 10 {
		return "标准打工人 💼", "朝九晚六标准型"
	}

	// 自由灵魂：晚到(>=09:30) + 正常走(<=20:00) + 短工时(<8h)
	if inTime >= 9*60+30 && outTime <= 20*60 && duration < 8 {
		return "自由灵魂 🌈", "来去自由的灵魂"
	}

	// 闪电侠：早到(<=08:30) + 早走(<=18:30) + 短工时(<8h)
	if inTime <= 8*60+30 && outTime <= 18*60+30 && duration < 8 {
		return "闪电侠 ⚡", "来得早走得也早"
	}

	// 摸鱼大师：任意时间 + 极短工时(<6h)
	if duration < 6 {
		return "摸鱼大师 🐟", "今天划水了"
	}

	// 默认：敬业达人
	return "敬业达人 💪", "认真工作的每一天"
}
