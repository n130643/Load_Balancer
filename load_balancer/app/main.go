package main

import (
	"encoding/json"
	constants "engine/load_balancer/constants"
	databaseLayer "engine/load_balancer/db"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   error  `json:"error"`
}

type row struct {
	Id       string `json:"id"`
	Url      string `json:"url"`
	Status   string `json:"status"`
	RRStatus string `json:"RRStatus"`
	Order    int    `json:"order"`
}

func Redirect(u string) string {

	switch u {
	case "127.0.0.1", "127.0.0.2": //"127.0.0.3", "127.0.0.4":
		return constants.Success

	case "127.0.0.3": //"127.0.0.5", "127.0.0.6", "127.0.0.7", "127.0.0.8":
		return constants.Failure
	default:
		return constants.Failure
	}
}

func parseBody(r *http.Request) (response, row) {

	bd, er := io.ReadAll(r.Body)
	var resp response = response{}
	if er != nil {
		resp = response{Status: http.StatusBadRequest, Message: "Failed to read body", Error: er}
		return resp, row{}
	} else if len(bd) == 0 {
		resp = response{Status: http.StatusBadRequest, Message: "Empty Body", Error: errors.New("empty Body")}
		return resp, row{}
	}

	var payload row
	er = json.Unmarshal(bd, &payload)
	if er != nil {
		resp = response{Status: http.StatusBadRequest, Message: "Failed to unmarshall request body", Error: er}
		return resp, row{}
	}
	return resp, payload
}

func RegisterRoute(payload *row) response {

	o, er := databaseLayer.GetMaxOrder()
	if er != nil {
		return response{Status: http.StatusInternalServerError, Message: er.Error()}
	}

	uuid := uuid.New()
	q := fmt.Sprintf("insert into load_balancers values('%s','%s','%s','%s',%d)", uuid.String(), payload.Url, "active", "active", o+1)
	fmt.Println(q)

	_, er = databaseLayer.ExeucteInserQuery(q)
	if er != nil {
		return response{Status: http.StatusInternalServerError, Message: er.Error()}
	}
	return response{Status: http.StatusOK, Message: "Successfully registered route"}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	resp, rw := parseBody(r)

	if resp.Error != nil {
		bs, _ := json.MarshalIndent(resp, "", "  ")
		w.Write(bs)
		return
	}

	resp = RegisterRoute(&rw)

	if resp.Error != nil {
		bs, _ := json.MarshalIndent(resp, "", "  ")
		w.Write(bs)
		return
	}
	resp = response{Status: http.StatusOK, Message: "Successfully registered route`"}
	bs, _ := json.MarshalIndent(resp, "", "  ")
	w.Write(bs)

}

func UpdateTablesPostRRPolicy(r row) response {

	result := Redirect(r.Url)

	uuid := uuid.New()
	q := fmt.Sprintf("insert into load_balancers_runs values('%s','%s','%s','%s')", uuid.String(), r.Id, time.Now().Format("2006-01-02 15:04:05"), result)
	_, er := databaseLayer.ExeucteInserQuery(q)
	if er != nil {
		return response{Status: http.StatusInternalServerError, Message: er.Error()}
	}

	q = fmt.Sprintf("update load_balancers set cur_rr_status = \"Inactive\" where id = '%s'", r.Id)
	_, er = databaseLayer.ExeucteInserQuery(q)
	if er != nil {
		return response{Status: http.StatusInternalServerError, Message: er.Error()}
	}

	mo, er := databaseLayer.GetMaxOrder()
	if er != nil {
		return response{Status: http.StatusInternalServerError, Message: er.Error()}
	}

	if r.Order == mo {
		q = "update load_balancers set cur_rr_status = \"active\""
		_, er = databaseLayer.ExeucteInserQuery(q)
		if er != nil {
			return response{Status: http.StatusInternalServerError, Message: er.Error()}
		}
	}
	if result == constants.Success {
		return response{Status: http.StatusOK, Message: fmt.Sprintf("successfully handled request by  %s", r.Url)}
	}
	return response{Status: http.StatusOK, Message: fmt.Sprintf("Failed to handle request by %s", r.Url)}

}

func ChooseRouteFromRRPolicy() (response, row) {

	var id, url, status, rrstatus string
	var order int
	minus50sec := time.Now().Add(-(time.Second * 60)).Format("2006-01-02 15:04:05")
	cur := time.Now().Format("2006-01-02 15:04:05")

	q := fmt.Sprintf(`select * from load_balancers where id not in 
			( select distinct url_id from load_balancers_runs where request_time 
				BETWEEN "%s" and "%s"  and response_status = "Failed" 
				group by url_id having count(*)>2
				) and cur_rr_status ="active"
		order by ll_order limit 1`, minus50sec, cur)

	rws, er := databaseLayer.Db.Query(q)
	if er != nil {
		return response{Status: http.StatusInternalServerError, Message: er.Error(), Error: er}, row{}
	}

	for rws.Next() {
		er = rws.Scan(&id, &url, &status, &rrstatus, &order)
		if er != nil {
			return response{Status: http.StatusInternalServerError, Message: er.Error(), Error: er}, row{}
		}
	}
	var resp response = response{}
	var rr row = row{}
	if id == "" {
		q = "update load_balancers set cur_rr_status = \"active\""
		_, er = databaseLayer.ExeucteInserQuery(q)
		if er != nil {
			return response{Status: http.StatusInternalServerError, Message: er.Error()}, row{}
		}
	} else {
		rr = row{Id: id, Order: order, Url: url}
		resp = UpdateTablesPostRRPolicy(rr)
		if resp.Error != nil {
			return resp, row{}
		}
	}

	return resp, rr

}

func routeRequest() response {

	resp, r := ChooseRouteFromRRPolicy()
	if resp.Error != nil {
		return resp
	} else if r.Url == "" {
		resp, r = ChooseRouteFromRRPolicy() // Retry to handle corner case.
		if resp.Error != nil {
			return resp
		} else if r.Url == "" {
			return response{Status: http.StatusServiceUnavailable, Message: "No Health Routes Available"}
		}
	}
	return resp
}

func proxyHanlder(w http.ResponseWriter, r *http.Request) {
	resp := routeRequest()

	bs, _ := json.MarshalIndent(resp, "", "  ")
	w.Write(bs)
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/urls/register", RegisterHandler).Methods("POST") //Handles routes registration
	r.HandleFunc("/proxy", proxyHanlder)                            // Distributes on RR manner
	fmt.Println("registered routes")
	er := http.ListenAndServe("0.0.0.0:8081", r)

	if er != nil {
		fmt.Println(er)
	}

}
