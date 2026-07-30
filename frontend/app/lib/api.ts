import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-node'

import { EmployeeService } from '../gen/employee/v1/employee_pb.ts'

// The BFF is the only Connect client; browsers never speak Connect directly.
const transport = createConnectTransport({
  baseUrl: process.env.BACKEND_URL ?? 'http://localhost:8080',
  httpVersion: '2',
})

export const employeeClient = createClient(EmployeeService, transport)
