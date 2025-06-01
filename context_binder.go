package ngebut

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// BindJSON unmarshals the JSON data from the request body into the provided object.
// It reads the request body, decodes the JSON, and populates the object.
// If the request body is nil or if unmarshaling fails, it returns an error.
// This method is typically used in route handlers to bind incoming JSON data to a struct.
// Parameters:
//   - obj: The object to unmarshal the JSON data into
//
// Returns:
//   - An error if the request body is nil or if unmarshaling fails
//   - nil if successful
//
// Example usage in a route handler:
//
//	func MyHandler(c *ngebut.Ctx) {
//		   var data MyDataType
//		   if err := c.BindJSON(&data); err != nil {
//		       c.Error(err)
//		       return
//		   }
//		   // Now data is populated with the JSON from the request body
//		   c.JSON(data)
//	}
func (c *Ctx) BindJSON(obj interface{}) error {
	if c.Request.Body == nil {
		return errors.New("request body is nil")
	}

	// Unmarshal the JSON data into the provided object
	if err := json.Unmarshal(c.Request.Body, obj); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// BindForm parses form data from the request and binds it to the provided object.
// It supports the following Content-Types:
// - application/x-www-form-urlencoded
// - multipart/form-data
// - text/plain (treated as URL-encoded)
// - empty Content-Type (treated as URL-encoded)
// The struct fields should be tagged with `form:"field_name"` to specify the form field name.
// If a field doesn't have a form tag, it will be skipped.
// Parameters:
//   - obj: The object to bind the form data to
//
// Returns:
//   - An error if parsing the form data fails or if the provided object is not a pointer to a struct
//   - nil if successful
//
// Example usage in a route handler:
//
//	func MyHandler(c *ngebut.Ctx) {
//	    var data MyDataType
//	    if err := c.BindForm(&data); err != nil {
//	        c.Error(err)
//	        return
//	    }
//	    // Now data is populated with the form data from the request
//	    c.JSON(data)
//	}
func (c *Ctx) BindForm(obj interface{}) error {
	// Check if the request has a body
	if c.Request.Body == nil {
		return errors.New("request body is nil")
	}

	// Check if obj is a pointer to a struct
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		return errors.New("obj must be a pointer to a struct")
	}

	// Parse the form data based on the Content-Type header
	contentType := c.Request.Header.Get("Content-Type")
	var values url.Values

	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		// Parse URL-encoded form data
		body := string(c.Request.Body)
		var err error
		values, err = url.ParseQuery(body)
		if err != nil {
			return fmt.Errorf("failed to parse form data: %w", err)
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form data
		// Create a new http.Request with the same body for parsing
		httpReq, err := http.NewRequest(c.Request.Method, c.Request.URL.String(), bytes.NewReader(c.Request.Body))
		if err != nil {
			return fmt.Errorf("failed to create request for multipart parsing: %w", err)
		}

		// Copy headers to ensure Content-Type with boundary is preserved
		for k, v := range *c.Request.Header {
			httpReq.Header[k] = v
		}

		// Parse the multipart form
		err = httpReq.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}

		values = httpReq.Form
	} else if contentType == "" || strings.HasPrefix(contentType, "text/plain") {
		// Handle plain form data or no content type (treat as URL-encoded)
		body := string(c.Request.Body)
		var err error
		values, err = url.ParseQuery(body)
		if err != nil {
			return fmt.Errorf("failed to parse form data: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported Content-Type for form binding: %s", contentType)
	}

	// Bind the form values to the struct fields
	objElem := objValue.Elem()
	objType := objElem.Type()

	for i := 0; i < objElem.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objElem.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the form tag
		formTag := field.Tag.Get("form")
		if formTag == "" {
			// Skip fields without a form tag
			continue
		}

		// Get the form value
		formValue := values.Get(formTag)
		if formValue == "" {
			// Skip empty values
			continue
		}

		// Set the field value based on its type
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(formValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(formValue, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as int: %w", formTag, err)
			}
			fieldValue.SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := strconv.ParseUint(formValue, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as uint: %w", formTag, err)
			}
			fieldValue.SetUint(uintValue)
		case reflect.Float32, reflect.Float64:
			floatValue, err := strconv.ParseFloat(formValue, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as float: %w", formTag, err)
			}
			fieldValue.SetFloat(floatValue)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(formValue)
			if err != nil {
				return fmt.Errorf("failed to parse %s as bool: %w", formTag, err)
			}
			fieldValue.SetBool(boolValue)
		default:
			// Skip unsupported types
			continue
		}
	}

	return nil
}
