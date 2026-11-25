package service

import (
	"context"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
)

// ApiKeyService lets the owner create, list and revoke agent-team members' keys.
type ApiKeyService struct {
	apiKeyRepository      domaininterface.IApiKeyRepository
	transactionRepository domaininterface.ITransactionRepository
	randomTokenProxy      domaininterface.IRandomTokenProxy
	clockProxy            domaininterface.IClockProxy
	calendar              domains.BusinessCalendarDomain
}

func NewApiKeyService(
	apiKeyRepository domaininterface.IApiKeyRepository,
	transactionRepository domaininterface.ITransactionRepository,
	randomTokenProxy domaininterface.IRandomTokenProxy,
	clockProxy domaininterface.IClockProxy,
	calendar domains.BusinessCalendarDomain,
) *ApiKeyService {
	return &ApiKeyService{
		apiKeyRepository:      apiKeyRepository,
		transactionRepository: transactionRepository,
		randomTokenProxy:      randomTokenProxy,
		clockProxy:            clockProxy,
		calendar:              calendar,
	}
}

// Create answers the full key this once; afterwards only its digest and display prefix exist.
func (apiKeyService *ApiKeyService) Create(
	executionContext context.Context, createDto dto.ApiKeyCreateDto,
) (dto.CreatedApiKeyDto, error) {
	issuedKey, validationError := domains.NewIssuedApiKeyDomain(createDto, apiKeyService.randomTokenProxy.NewToken(),
		apiKeyService.clockProxy.Now(), apiKeyService.calendar)
	if validationError != nil {
		return dto.CreatedApiKeyDto{}, validationError
	}

	var createdDto dto.CreatedApiKeyDto
	createError := apiKeyService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			isTaken, checkError := apiKeyService.apiKeyRepository.HasActiveMemberName(
				transactionContext, issuedKey.MemberNameKey())
			if checkError != nil {
				return checkError
			}
			if isTaken {
				return domains.Conflict("memberNameInUse", "這個成員名稱已有啟用中的金鑰")
			}

			created, storeError := apiKeyService.apiKeyRepository.Create(transactionContext, issuedKey.ToEntity())
			if storeError != nil {
				return storeError
			}
			createdDto = dto.CreatedApiKeyDto{
				ApiKeyDto: domains.NewApiKeyDomain(created, apiKeyService.calendar).ToDto(),
				Key:       issuedKey.Key(),
			}

			return nil
		})
	if createError != nil {
		return dto.CreatedApiKeyDto{}, createError
	}

	return createdDto, nil
}

func (apiKeyService *ApiKeyService) List(executionContext context.Context) ([]dto.ApiKeyDto, error) {
	apiKeys, listError := apiKeyService.apiKeyRepository.List(executionContext)
	if listError != nil {
		return nil, listError
	}

	apiKeyDtos := make([]dto.ApiKeyDto, 0, len(apiKeys))
	for _, apiKey := range apiKeys {
		apiKeyDtos = append(apiKeyDtos, domains.NewApiKeyDomain(apiKey, apiKeyService.calendar).ToDto())
	}

	return apiKeyDtos, nil
}

// Revoke is final; revoking a key already revoked answers it unchanged.
func (apiKeyService *ApiKeyService) Revoke(executionContext context.Context, apiKeyID string) (dto.ApiKeyDto, error) {
	var revokedDto dto.ApiKeyDto
	revokeError := apiKeyService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			apiKey, isFound, findError := apiKeyService.apiKeyRepository.FindByID(transactionContext, apiKeyID)
			if findError != nil {
				return findError
			}
			if !isFound {
				return domains.NotFound("apiKeyNotFound", "找不到這把金鑰")
			}

			if domains.NewApiKeyDomain(apiKey, apiKeyService.calendar).IsActive() {
				if storeError := apiKeyService.apiKeyRepository.Revoke(
					transactionContext, apiKeyID, apiKeyService.clockProxy.Now()); storeError != nil {
					return storeError
				}
				apiKey, _, findError = apiKeyService.apiKeyRepository.FindByID(transactionContext, apiKeyID)
				if findError != nil {
					return findError
				}
			}
			revokedDto = domains.NewApiKeyDomain(apiKey, apiKeyService.calendar).ToDto()

			return nil
		})
	if revokeError != nil {
		return dto.ApiKeyDto{}, revokeError
	}

	return revokedDto, nil
}
