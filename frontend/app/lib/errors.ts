import { Code, ConnectError } from '@connectrpc/connect'

const userFacing = new Set([
  Code.InvalidArgument,
  Code.NotFound,
  Code.AlreadyExists,
  Code.FailedPrecondition,
])

// domainErrorMessage returns the backend's message for errors a user can act
// on and rethrows everything else (infrastructure failures stay 500s).
export function domainErrorMessage(error: unknown): string {
  if (error instanceof ConnectError && userFacing.has(error.code)) {
    return error.rawMessage
  }
  throw error
}
