/*
 * API для взаимодействия с STM32MP1
 *
 * Данное API чото гдето зочемто нужно, не очень понятно. Но пусть будет, что мешает штоли?
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type RespFileElem struct {

	Results *RespFileElemResults `json:"results,omitempty"`
	// Three possible statuses:   * `OK`: No errors occurred.  * `INVALID_REQUEST`: Some parameters are missing or invalid.  * `EXECUTE_ERROR`: No or wrong responce from Power Control Block.  * `UNKNOWN_ERROR`: The request could not be processed due to a server error. The request may succeed if you try again.
	Status string `json:"status,omitempty"`
}
