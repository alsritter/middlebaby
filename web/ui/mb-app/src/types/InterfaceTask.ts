export interface InterfaceTask {
  protocol: string
  serviceName: string
  serviceMethod: string
  serviceDescription: string
  servicePath: string
  serviceProtoFile: string
  setup: Setup[]
  mocks: Mock[]
  teardown: Teardown[]
  cases: Case[]
}

export interface Setup {
  typeName: string
  commands: string[]
}

export interface Mock {
  request: Request
  response: Response
}

export interface Request {
  protocol: string
  method: string
  host: string
  path: string
  header: any
  params: Params
  body: any
}

export interface Params {
  age: string
  name: string
}

export interface Response {
  status: number
  header: any
  body: string
  trailer: any
  delay: Delay
}


export interface Delay {
  delay: number
  offset: number
}

export interface Teardown {
  typeName: string
  commands: string[]
}

export interface Case {
  name: string
  description: string
  setup: Setup[]
  mocks: Mock[]
  request: Request
  assert: Assert
  teardown: Teardown[]
}

export interface Query {
  a?: string[]
  b?: string[]
}

export interface Assert {
  Response: AssertResponse
  otherAsserts?: OtherAssert[]
}

export interface AssertResponse {
  header: any
  body: any
}

export interface OtherAssert {
  typeName: string
  actual: string
  expected: any
}
