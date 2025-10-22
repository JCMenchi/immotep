import { beforeAll, afterEach, afterAll } from 'vitest'
import { server } from './mock/httpserver.js'
import '@testing-library/jest-dom'
import { cleanup } from '@testing-library/react'

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => {
    server.close()
    cleanup()
})
