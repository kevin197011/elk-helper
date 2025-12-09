# Change: Add Cleanup Task Execution Status

## Why

Users need visibility into cleanup task execution status to monitor whether automatic cleanup tasks are running successfully. Currently, there's no way to know if cleanup tasks have executed, when they last ran, or whether they succeeded or failed.

## What Changes

- Added execution status tracking fields to CleanupConfig model:
  - `last_execution_status`: Status of last execution ("success", "failed", "never")
  - `last_execution_time`: Timestamp of last execution
  - `last_execution_result`: Result description (deletion count or error message)

- Updated cleanup task execution logic to record status after each run (both automatic and manual)

- Added UI display on cleanup config page showing last execution status with visual indicators

## Impact

- Affected specs: `specs/system-config/spec.md`
- Affected code:
  - `backend/internal/models/system_config.go` - Model changes
  - `backend/internal/service/systemconfig/system_config.go` - Service method for status updates
  - `backend/internal/worker/scheduler/scheduler.go` - Automatic cleanup status tracking
  - `backend/internal/api/handlers/system_config_handler.go` - Manual cleanup status tracking
  - `frontend/src/services/api.ts` - TypeScript interface updates
  - `frontend/src/pages/CleanupConfigPage.tsx` - UI display

