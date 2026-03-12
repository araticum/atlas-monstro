package v1

import (
    "net/http"
    "strconv"

    "code.vikunja.io/api/pkg/db"
    "code.vikunja.io/api/pkg/models"
    auth2 "code.vikunja.io/api/pkg/modules/auth"
    "github.com/labstack/echo/v5"
)

func GetTaskActivities(c *echo.Context) error {
    taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.").Wrap(err)
    }

    s := db.NewSession()
    defer s.Close()

    a, err := auth2.GetAuthFromClaims(c)
    if err != nil {
        return err
    }

    activities, err := models.GetTaskActivities(s, taskID, a)
    if err != nil {
        return err
    }

    return c.JSON(http.StatusOK, activities)
}
