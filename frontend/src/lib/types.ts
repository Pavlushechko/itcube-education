// src/lib/types.ts

export type Role = 'user' | 'moderator' | 'admin'

export type Program = {
  ID: string
  Title: string
  Description: string
  Status?: string
  CreatedAt?: string
}

export type Group = {
  ID: string
  ProgramID: string
  CohortID: string
  Title: string
  Capacity: number
  IsOpen: boolean
  RequiresInterview: boolean
  CreatedAt?: string
}

export type ProgramWithGroups = {
  Program: Program
  Groups: Group[]
}

export type Application = {
  ID: string
  UserID: string
  GroupID: string
  Status: string
  Comment: string
  CreatedAt: string
  UpdatedAt: string
}

export type Material = {
  ID: string
  GroupID: string
  Type: string
  Title: string
  Content: string
  CreatedBy: string
  CreatedAt: string
}
