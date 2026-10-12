import Redis from 'ioredis';

const globalForRedis = globalThis as unknown as {
  redis?: Redis;
};

const redisUrl = process.env.REDIS_URL || 'redis://localhost:6379';

export const redisClient =
  globalForRedis.redis ||
  new Redis(redisUrl, {
    maxRetriesPerRequest: 3,
    retryStrategy(times) {
      const delay = Math.min(times * 50, 2000);
      return delay;
    },
    lazyConnect: true,
  });

if (process.env.NODE_ENV !== 'production') {
  globalForRedis.redis = redisClient;
}

/**
 * Checks if Redis is connected and responsive.
 */
export async function checkRedisConnection(): Promise<boolean> {
  try {
    const res = await redisClient.ping();
    return res === 'PONG';
  } catch {
    return false;
  }
}
