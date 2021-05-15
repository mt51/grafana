package api

import (
	"errors"
	"net/http"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/metrics"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

// GET /api/org
func GetOrgCurrent(c *models.ReqContext) response.Response {
	return getOrgHelper(c.OrgId)
}

// GET /api/orgs/:orgId
func GetOrgByID(c *models.ReqContext) response.Response {
	return getOrgHelper(c.ParamsInt64(":orgId"))
}

// Get /api/orgs/name/:name
func (hs *HTTPServer) GetOrgByName(c *models.ReqContext) response.Response {
	org, err := hs.SQLStore.GetOrgByName(c.Params(":name"))
	if err != nil {
		if errors.Is(err, models.ErrOrgNotFound) {
			return response.Error(http.StatusNotFound, "Organization not found", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to get organization", err)
	}
	result := models.OrgDetailsDTO{
		Id:   org.Id,
		Name: org.Name,
		Address: models.Address{
			Address1: org.Address1,
			Address2: org.Address2,
			City:     org.City,
			ZipCode:  org.ZipCode,
			State:    org.State,
			Country:  org.Country,
		},
	}

	return response.JSON(http.StatusOK, &result)
}

func getOrgHelper(orgID int64) response.Response {
	query := models.GetOrgByIdQuery{Id: orgID}

	if err := bus.Dispatch(&query); err != nil {
		if errors.Is(err, models.ErrOrgNotFound) {
			return response.Error(http.StatusNotFound, "Organization not found", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to get organization", err)
	}

	org := query.Result
	result := models.OrgDetailsDTO{
		Id:   org.Id,
		Name: org.Name,
		Address: models.Address{
			Address1: org.Address1,
			Address2: org.Address2,
			City:     org.City,
			ZipCode:  org.ZipCode,
			State:    org.State,
			Country:  org.Country,
		},
	}

	return response.JSON(http.StatusOK, &result)
}

// POST /api/orgs
func CreateOrg(c *models.ReqContext, cmd models.CreateOrgCommand) response.Response {
	if !c.IsSignedIn || (!setting.AllowUserOrgCreate && !c.IsGrafanaAdmin) {
		return response.Error(http.StatusForbidden, "Access denied", nil)
	}

	cmd.UserId = c.UserId
	if err := bus.Dispatch(&cmd); err != nil {
		if errors.Is(err, models.ErrOrgNameTaken) {
			return response.Error(http.StatusConflict, "Organization name taken", err)
		}
		return response.Error(http.StatusInternalServerError, "Failed to create organization", err)
	}

	metrics.MApiOrgCreate.Inc()

	return response.JSON(http.StatusOK, &util.DynMap{
		"orgId":   cmd.Result.Id,
		"message": "Organization created",
	})
}

// PUT /api/org
func UpdateOrgCurrent(c *models.ReqContext, form dtos.UpdateOrgForm) response.Response {
	return updateOrgHelper(form, c.OrgId)
}

// PUT /api/orgs/:orgId
func UpdateOrg(c *models.ReqContext, form dtos.UpdateOrgForm) response.Response {
	return updateOrgHelper(form, c.ParamsInt64(":orgId"))
}

func updateOrgHelper(form dtos.UpdateOrgForm, orgID int64) response.Response {
	cmd := models.UpdateOrgCommand{Name: form.Name, OrgId: orgID}
	if err := bus.Dispatch(&cmd); err != nil {
		if errors.Is(err, models.ErrOrgNameTaken) {
			return response.Error(http.StatusBadRequest, "Organization name taken", err)
		}
		return response.Error(http.StatusInternalServerError, "Failed to update organization", err)
	}

	return response.Success("Organization updated")
}

// PUT /api/org/address
func UpdateOrgAddressCurrent(c *models.ReqContext, form dtos.UpdateOrgAddressForm) response.Response {
	return updateOrgAddressHelper(form, c.OrgId)
}

// PUT /api/orgs/:orgId/address
func UpdateOrgAddress(c *models.ReqContext, form dtos.UpdateOrgAddressForm) response.Response {
	return updateOrgAddressHelper(form, c.ParamsInt64(":orgId"))
}

func updateOrgAddressHelper(form dtos.UpdateOrgAddressForm, orgID int64) response.Response {
	cmd := models.UpdateOrgAddressCommand{
		OrgId: orgID,
		Address: models.Address{
			Address1: form.Address1,
			Address2: form.Address2,
			City:     form.City,
			State:    form.State,
			ZipCode:  form.ZipCode,
			Country:  form.Country,
		},
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to update org address", err)
	}

	return response.Success("Address updated")
}

// GET /api/orgs/:orgId
func DeleteOrgByID(c *models.ReqContext) response.Response {
	if err := bus.Dispatch(&models.DeleteOrgCommand{Id: c.ParamsInt64(":orgId")}); err != nil {
		if errors.Is(err, models.ErrOrgNotFound) {
			return response.Error(http.StatusNotFound, "Failed to delete organization. ID not found", nil)
		}
		return response.Error(http.StatusInternalServerError, "Failed to update organization", err)
	}
	return response.Success("Organization deleted")
}

func SearchOrgs(c *models.ReqContext) response.Response {
	perPage := c.QueryInt("perpage")
	if perPage <= 0 {
		perPage = 1000
	}

	page := c.QueryInt("page")

	query := models.SearchOrgsQuery{
		Query: c.Query("query"),
		Name:  c.Query("name"),
		Page:  page,
		Limit: perPage,
	}

	if err := bus.Dispatch(&query); err != nil {
		return response.Error(http.StatusInternalServerError, "Failed to search orgs", err)
	}

	return response.JSON(http.StatusOK, query.Result)
}
