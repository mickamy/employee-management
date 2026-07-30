import { create } from '@bufbuild/protobuf'

import { DateSchema, type Date as ProtoDate } from '../gen/google/type/date_pb.ts'

// toProtoDate parses the wire format of <input type="date">, "YYYY-MM-DD".
export function toProtoDate(value: string): ProtoDate {
  const [year, month, day] = value.split('-').map(Number)
  return create(DateSchema, { year, month, day })
}

export function formatDate(date?: ProtoDate): string {
  if (!date || date.year === 0) {
    return ''
  }
  const month = String(date.month).padStart(2, '0')
  const day = String(date.day).padStart(2, '0')
  return `${date.year}-${month}-${day}`
}
