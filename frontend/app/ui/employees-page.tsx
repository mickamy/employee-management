import type { Handle } from 'remix/ui'

import type { Employee } from '../gen/employee/v1/employee_pb.ts'
import { formatDate } from '../lib/dates.ts'
import { routes } from '../routes.ts'
import { Document } from './document.tsx'

export interface EmployeesPageProps {
  employees: Employee[]
  nextPageToken: string
  error?: string
}

export function EmployeesPage(handle: Handle<EmployeesPageProps>) {
  return () => {
    const { employees, nextPageToken, error } = handle.props

    return (
      <Document title="Employees">
        <main>
          <h1>Employees</h1>

          <table>
            <thead>
              <tr>
                <th>Code</th>
                <th>Name</th>
                <th>Email</th>
                <th>Hired on</th>
              </tr>
            </thead>
            <tbody>
              {employees.map((employee) => (
                <tr>
                  <td>{employee.code}</td>
                  <td>{employee.name}</td>
                  <td>{employee.email}</td>
                  <td>{formatDate(employee.hiredOn)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {nextPageToken !== '' && (
            <p>
              <a href={`${routes.employees.index.href()}?page_token=${nextPageToken}`}>Next page</a>
            </p>
          )}

          <h2>Hire an employee</h2>
          {error && <p role="alert">{error}</p>}
          <form method="post" action={routes.employees.action.href()}>
            <p>
              <label>
                Code <input name="code" required />
              </label>
            </p>
            <p>
              <label>
                Name <input name="name" required />
              </label>
            </p>
            <p>
              <label>
                Email <input name="email" type="email" required />
              </label>
            </p>
            <p>
              <label>
                Hired on <input name="hired_on" type="date" required />
              </label>
            </p>
            <button type="submit">Hire</button>
          </form>
        </main>
      </Document>
    )
  }
}
