package exporter

type OutputFormat int

const (
	OutputTelegramList OutputFormat = iota
	OutputExcel
)

func ChooseFormat(participantsCount int) OutputFormat {
	if participantsCount < 50 {
		return OutputTelegramList
	}
	return OutputExcel
}
