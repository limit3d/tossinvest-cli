package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/junghoonkye/toss-investment-cli/internal/domain"
)

func WriteOrders(w io.Writer, format Format, orders []domain.Order) error {
	switch format {
	case FormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(orders)
	case FormatCSV:
		writer := csv.NewWriter(w)
		if err := writer.Write([]string{"count"}); err != nil {
			return err
		}
		if err := writer.Write([]string{strconv.Itoa(len(orders))}); err != nil {
			return err
		}
		writer.Flush()
		return writer.Error()
	case FormatTable:
		if len(orders) == 0 {
			_, err := fmt.Fprintln(w, "No pending orders")
			return err
		}
		_, err := fmt.Fprintf(w, "Pending Orders: %d\n", len(orders))
		return err
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
