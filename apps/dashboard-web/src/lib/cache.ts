import { redisClient } from './redis';

/**
 * Retrieves cached JSON data from Redis, or fetches it via the provided getter function,
 * caches it with a TTL (default 5 minutes), and returns the result.
 */
export async function getCachedData<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttlSeconds: number = 300
): Promise<T> {
  try {
    const cached = await redisClient.get(key);
    if (cached) {
      return JSON.parse(cached) as T;
    }
  } catch (err) {
    console.warn(`[Redis Cache Warning] Failed to read key ${key}:`, err);
  }

  const freshData = await fetcher();

  try {
    await redisClient.set(key, JSON.stringify(freshData), 'EX', ttlSeconds);
  } catch (err) {
    console.warn(`[Redis Cache Warning] Failed to write key ${key}:`, err);
  }

  return freshData;
}

/**
 * Invalidates / deletes a cache key from Redis.
 */
export async function invalidateCache(key: string): Promise<void> {
  try {
    await redisClient.del(key);
  } catch (err) {
    console.warn(`[Redis Cache Warning] Failed to delete key ${key}:`, err);
  }
}
