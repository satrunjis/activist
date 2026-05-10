import { z } from "zod";

export const roleFormSchema = z.object({
  name: z.string().trim().min(1, "Укажите название роли.").max(100, "Название роли должно быть не длиннее 100 символов."),
  permissions: z.array(z.string())
});

export type RoleFormValues = z.infer<typeof roleFormSchema>;
