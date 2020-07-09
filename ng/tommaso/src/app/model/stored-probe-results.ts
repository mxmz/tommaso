export interface StoredProbeResults {
    type: string
    args: string[]
    source: string
    status: string
    time: string
    elapsed: number
    comment: string
    description: string
    expect_failure: boolean
    pass: boolean
}
