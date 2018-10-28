package Clock_Trigger

import (
	"net/http"
	"time"
)

func main() {

	for range time.Tick(20 * time.Second) {
		http.Get("https://igcinfo-imt2681.herokuapp.com/paragliding/admin/api/webhook")
	}
}
