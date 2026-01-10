package exporter

import (
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strings"
	"time"
)

type Options struct {
	ExportedAt time.Time
}

func ExportExcel(result ParticipantsResult, opt Options) ([]byte, error) {
	exportedAt := opt.ExportedAt
	if exportedAt.IsZero() {
		exportedAt = time.Now()
	}

	f := excelize.NewFile()
	const (
		sParticipants = "Participants"
		sMentions     = "Mentions"
		sChannels     = "Channels"
	)

	if _, err := f.NewSheet(sParticipants); err != nil {
		return nil, err
	}
	if _, err := f.NewSheet(sMentions); err != nil {
		return nil, err
	}
	if _, err := f.NewSheet(sChannels); err != nil {
		return nil, err
	}

	headers := []string{
		"Дата экспорта",
		"Username",
		"Имя и фамилия",
		"Описание",
		"Дата регистрации",
		"Наличие канала",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	dateStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 22,
	})

	if err := writePeopleSheet(f, sParticipants, headers, result.Participants, exportedAt, headerStyle, dateStyle); err != nil {
		return nil, err
	}
	if err := writePeopleSheet(f, sMentions, headers, result.Mentions, exportedAt, headerStyle, dateStyle); err != nil {
		return nil, err
	}
	if err := writeChannelsSheet(f, sChannels, exportedAt, result.Channels, headerStyle, dateStyle); err != nil {
		return nil, err
	}

	_ = f.DeleteSheet("Sheet1")

	i, err := f.GetSheetIndex(sParticipants)
	if err != nil {
		return nil, err
	}

	f.SetActiveSheet(i)

	var buf bytes.Buffer
	if err = f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writePeopleSheet(
	f *excelize.File,
	sheet string,
	headers []string,
	people []Participant,
	exportedAt time.Time,
	headerStyle int,
	dateStyle int,
) error {
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	row := 2
	for _, p := range people {
		if p.IsDeleted {
			continue
		}
		if strings.TrimSpace(p.Username) == "" && strings.TrimSpace(p.FirstName+p.LastName) == "" {
			continue
		}

		values := make([]any, 0, len(headers))
		values = append(values, exportedAt)

		username := strings.TrimSpace(p.Username)
		if username != "" && !strings.HasPrefix(username, "@") {
			username = "@" + username
		}
		values = append(values, username)

		fullName := strings.TrimSpace(strings.TrimSpace(p.FirstName) + " " + strings.TrimSpace(p.LastName))
		values = append(values, fullName)

		values = append(values, p.Bio)

		if p.RegisteredAt != nil {
			values = append(values, *p.RegisteredAt)
		} else {
			values = append(values, "")
		}

		if p.HasChannel {
			values = append(values, "да")
		} else {
			values = append(values, "нет")
		}

		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				return err
			}
			if col == 0 || col == 4 {
				_ = f.SetCellStyle(sheet, cell, cell, dateStyle)
			}
		}
		row++
	}

	_ = f.SetColWidth(sheet, "A", "A", 20)
	_ = f.SetColWidth(sheet, "B", "B", 18)
	_ = f.SetColWidth(sheet, "C", "C", 22)
	_ = f.SetColWidth(sheet, "D", "D", 40)
	_ = f.SetColWidth(sheet, "E", "E", 20)
	_ = f.SetColWidth(sheet, "F", "F", 16)

	return nil
}

func writeChannelsSheet(
	f *excelize.File,
	sheet string,
	exportedAt time.Time,
	channels []string,
	headerStyle int,
	dateStyle int,
) error {
	if err := f.SetCellValue(sheet, "A1", "Дата экспорта"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "B1", "Channel"); err != nil {
		return err
	}
	_ = f.SetCellStyle(sheet, "A1", "B1", headerStyle)

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	row := 2
	for _, ch := range channels {
		ch = strings.TrimSpace(ch)
		if ch == "" {
			continue
		}
		a, _ := excelize.CoordinatesToCellName(1, row)
		b, _ := excelize.CoordinatesToCellName(2, row)
		_ = f.SetCellValue(sheet, a, exportedAt)
		_ = f.SetCellStyle(sheet, a, a, dateStyle)
		_ = f.SetCellValue(sheet, b, ch)
		row++
	}

	_ = f.SetColWidth(sheet, "A", "A", 20)
	_ = f.SetColWidth(sheet, "B", "B", 40)

	if row == 2 {
		_ = f.SetCellValue(sheet, "A2", exportedAt)
		_ = f.SetCellStyle(sheet, "A2", "A2", dateStyle)
		_ = f.SetCellValue(sheet, "B2", "(каналы не найдены)")
	}

	return nil
}

func ValidateExcelBytes(xlsx []byte) error {
	f, err := excelize.OpenReader(bytes.NewReader(xlsx))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	required := []string{"Participants", "Mentions", "Channels"}
	for _, s := range required {
		if _, err = f.GetSheetIndex(s); err != nil {
			return fmt.Errorf("missing sheet %q", s)
		}
	}
	return nil
}
