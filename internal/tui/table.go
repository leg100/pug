package tui

//import (
//	tea "github.com/charmbracelet/bubbletea"
//	wrapped "github.com/evertras/bubble-table/table"
//	"github.com/leg100/pug/internal/resource"
//)

//type table[T any] struct {
//	wrapped.Model
//}
//
//func (m table[T]) Init() tea.Cmd {
//	return nil
//}
//
//func (m table[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	switch msg := msg.(type) {
//	case resource.Event[T]:
//		switch msg.Type {
//		case resource.CreatedEvent:
//			// Insert new at top
//			m.Model.WithRows(
//				append([]wrapped.Row{newRunRow(msg.Payload)}, m.table.Rows()...),
//			)
//		case resource.UpdatedEvent:
//			i := m.findRow(msg.Payload.ID)
//			if i < 0 {
//				// TODO: log error
//				return m, nil
//			}
//			// remove row
//			rows := append(m.table.Rows()[:i], m.table.Rows()[i+1:]...)
//			// add to top
//			m.table.SetRows(
//				append([]table.Row{newRunRow(msg.Payload)}, rows...),
//			)
//		case resource.DeletedEvent:
//			i := m.findRow(msg.Payload.ID)
//			if i < 0 {
//				// TODO: log error
//				return m, nil
//			}
//			// remove row
//			m.table.SetRows(
//				append(m.table.Rows()[:i], m.table.Rows()[i+1:]...),
//			)
//		}
//	}
//}
