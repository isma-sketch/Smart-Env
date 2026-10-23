package ui

import "github.com/charmbracelet/lipgloss"

var (
	Title    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	Subtitle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	OK       = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	Fail     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	Warn     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	Info     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	Bold     = lipgloss.NewStyle().Bold(true)
	Dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	Box      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

func PrintTitle(s string) { println(Title.Render(s)) }

func Success(s string) { println(OK.Render("✓ " + s)) }
