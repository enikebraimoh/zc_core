package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// Add a report
func AddReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var report Report

	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(orgId)
	
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	org_collection, member_collection := "organizations", "members"
	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"_id": objId})
	if orgDoc == nil {
		fmt.Printf("organization with id %s doesn't exist!", orgId)
		utils.GetError(errors.New("organization with id "+ orgId + " doesn't exist!"), http.StatusBadRequest, w)
		return
	}

	report.Organization = orgId
	report.Date = time.Now()

	if !utils.IsValidEmail(report.ReporterEmail) {
		utils.GetError(fmt.Errorf("invalid email format : %s", report.ReporterEmail), http.StatusBadRequest, w)
		return
	}

	// check that reporter is in the organization
	reporterDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": report.ReporterEmail})
	if reporterDoc == nil {
		utils.GetError(errors.New("reporter must be a member of this organization"), http.StatusBadRequest, w)
		return
	}

	if !utils.IsValidEmail(report.ReporteeEmail) {
		utils.GetError(fmt.Errorf("invalid email format : %s", report.ReporteeEmail), http.StatusBadRequest, w)
		return
	}

	// check that reportee is in the organization
	reporteeDoc, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": report.ReporteeEmail})
	if reporteeDoc == nil {
		utils.GetError(errors.New("reportee must be a member of this organization"), http.StatusBadRequest, w)
		return
	}

	if report.Organization == "" {
		utils.GetError(errors.New("organization id required"), http.StatusBadRequest, w)
		return
	}

	if report.Subject == "" {
		utils.GetError(errors.New("report's subject required"), http.StatusBadRequest, w)
		return
	}

	if report.Body == "" {
		utils.GetError(errors.New("report's body required"), http.StatusBadRequest, w)
		return
	}	

	
	var reportMap map[string]interface{}
	utils.ConvertStructure(report, &reportMap)

	save, err := utils.CreateMongoDbDoc(ReportCollectionName, reportMap)
	if err != nil {
		fmt.Println(err)
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("report created", save, w)
}

// Get a report
func GetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]

	reportId := mux.Vars(r)["report_id"]
	reportObjId, err := primitive.ObjectIDFromHex(reportId)
	
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	doc, _ := utils.GetMongoDbDoc(ReportCollectionName, bson.M{"organization_id": orgId, "_id": reportObjId})

	if doc == nil {
		utils.GetError(fmt.Errorf("report %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var report Report
	utils.ConvertStructure(doc, &report)

	utils.GetSuccess("report  retrieved successfully", report, w)
}

// Get reports
func GetReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]

	doc, _ := utils.GetMongoDbDocs(ReportCollectionName, bson.M{"organization_id": orgId})

	/* if doc == nil {
		utils.GetError(fmt.Errorf("reports %s not found", orgId), http.StatusNotFound, w)
		return
	}

	var report []Report */

	var report []Report = []Report{}
	// var re

	if doc == nil {
		utils.GetSuccess("no report has been added yet", report, w)
		return
	}
	
	utils.ConvertStructure(doc, &report)

	utils.GetSuccess("report  retrieved successfully", report, w)
}