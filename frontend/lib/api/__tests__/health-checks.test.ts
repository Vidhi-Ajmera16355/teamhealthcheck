import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the shared client module so we can assert on apiRequest calls without
// going through authenticatedFetch (which pulls in cookies/localStorage auth state).
vi.mock('@/lib/api/client', () => ({
  API_BASE_URL: '',
  APIRequestError: class APIRequestError extends Error {},
  apiRequest: vi.fn(),
  handleResponse: vi.fn(),
}));

import { apiRequest, handleResponse } from '@/lib/api/client';
import { saveDraft, getDraft } from '../health-checks';

const mockApiRequest = apiRequest as unknown as ReturnType<typeof vi.fn>;
const mockHandleResponse = handleResponse as unknown as ReturnType<typeof vi.fn>;

beforeEach(() => {
  // vi.restoreAllMocks() only restores vi.spyOn spies to their original implementation;
  // these are plain vi.fn() mocks from the module factory, so reset them explicitly.
  mockApiRequest.mockReset();
  mockHandleResponse.mockReset();
});

describe('saveDraft', () => {
  it('PUTs the draft payload to /api/v1/health-checks/draft and returns the parsed record', async () => {
    const payload = {
      teamId: 'team1',
      userId: 'user1',
      surveyType: 'individual' as const,
      assessmentPeriod: '2026 - 1st Half',
      currentDimension: 2,
      responses: [{ dimensionId: 'mission', score: 3 as const, trend: 'stable' as const, comment: '' }],
      clientUpdatedAt: 1700000000000,
    };
    const record = { id: 'draft-user1-team1-individual', ...payload };

    mockApiRequest.mockResolvedValue({ ok: true });
    mockHandleResponse.mockResolvedValue(record);

    const result = await saveDraft(payload);

    expect(mockApiRequest).toHaveBeenCalledWith('/api/v1/health-checks/draft', {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
    expect(result).toEqual(record);
  });
});

describe('getDraft', () => {
  it('returns null when the server responds 404 (no draft exists)', async () => {
    mockApiRequest.mockResolvedValue({ status: 404, ok: false });

    const result = await getDraft('team1', 'user1', 'individual');

    expect(result).toBeNull();
    expect(mockHandleResponse).not.toHaveBeenCalled();
    expect(mockApiRequest).toHaveBeenCalledWith(
      '/api/v1/health-checks/draft?teamId=team1&userId=user1&surveyType=individual'
    );
  });

  it('returns the parsed draft when one exists', async () => {
    const record = {
      id: 'draft-user1-team1-individual',
      teamId: 'team1',
      userId: 'user1',
      surveyType: 'individual',
      assessmentPeriod: '2026 - 1st Half',
      currentDimension: 1,
      responses: [],
      clientUpdatedAt: 1700000000000,
    };
    mockApiRequest.mockResolvedValue({ status: 200, ok: true });
    mockHandleResponse.mockResolvedValue(record);

    const result = await getDraft('team1', 'user1');

    expect(result).toEqual(record);
  });

  it('defaults surveyType to individual when not provided', async () => {
    mockApiRequest.mockResolvedValue({ status: 404, ok: false });

    await getDraft('team1', 'user1');

    expect(mockApiRequest).toHaveBeenCalledWith(
      '/api/v1/health-checks/draft?teamId=team1&userId=user1&surveyType=individual'
    );
  });
});
