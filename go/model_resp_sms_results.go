/*
 * API для взаимодействия с STM32MP1
 *
 * Данное API чото гдето зочемто нужно, не очень понятно. Но пусть будет, что мешает штоли?
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type RespSmsResults struct {
	// Номер телефона
	Phone string `json:"phone"`
	// СМС сообщение
	Message string `json:"message"`
}
