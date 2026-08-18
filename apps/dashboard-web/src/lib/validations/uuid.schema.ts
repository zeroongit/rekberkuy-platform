import { z } from 'zod';

export const uuidParamSchema = z.object({
  id: z.string().uuid('ID sumber daya harus berupa UUID v4 yang sah (mencegah IDOR & enumeration)'),
});

export type UuidParamInput = z.infer<typeof uuidParamSchema>;
