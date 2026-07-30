import { redirect } from 'remix/response/redirect'
import { createController } from 'remix/router'

import { employeeClient } from '../lib/api.ts'
import { toProtoDate } from '../lib/dates.ts'
import { domainErrorMessage } from '../lib/errors.ts'
import { routes } from '../routes.ts'
import { EmployeesPage } from '../ui/employees-page.tsx'

const pageSize = 20

export default createController(routes.employees, {
  actions: {
    async index(context) {
      const pageToken = context.url.searchParams.get('page_token') ?? ''
      const { employees, nextPageToken } = await employeeClient.listEmployees({
        pageSize,
        pageToken,
      })
      return context.render(<EmployeesPage employees={employees} nextPageToken={nextPageToken} />)
    },

    async action(context) {
      const form = await context.request.formData()
      try {
        await employeeClient.hireEmployee({
          code: String(form.get('code') ?? ''),
          name: String(form.get('name') ?? ''),
          email: String(form.get('email') ?? ''),
          hiredOn: toProtoDate(String(form.get('hired_on') ?? '')),
        })
      } catch (error) {
        const message = domainErrorMessage(error)
        const { employees, nextPageToken } = await employeeClient.listEmployees({ pageSize })
        return context.render(
          <EmployeesPage employees={employees} nextPageToken={nextPageToken} error={message} />,
          { status: 422 },
        )
      }
      return redirect(routes.employees.index.href(), 303)
    },
  },
})
