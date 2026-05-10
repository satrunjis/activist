import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ProfilePage } from "./ProfilePage";
import type { UserProfile } from "../../shared/api/types";
import { ToastProvider } from "../../shared/ui/feedback";

const mockRequest = vi.fn();

vi.mock("../../shared/api/client", () => ({
  request: (...args: unknown[]) => mockRequest(...args)
}));

const baseProfile: UserProfile = {
  id: "user-1",
  login: "ivanov",
  first_name: "Иван",
  last_name: "Иванов",
  middle_name: "",
  gradebook_number: "12345",
  group_number: "A-01",
  institute: "ИТ",
  birth_date: "2001-02-03T00:00:00Z",
  phone: "",
  social_links: [],
  about: "",
  memberships: []
};

function renderProfile(profile: UserProfile = baseProfile) {
  mockRequest.mockImplementation(async (path: string, options?: { method?: string; body?: unknown }) => {
    if (options?.method === "PATCH") {
      return { ...profile, ...(options.body as Partial<UserProfile>) };
    }
    if (path.endsWith("/memberships")) {
      return { items: [] };
    }
    return profile;
  });

  return render(
    <ToastProvider>
      <ProfilePage userId="user-1" />
    </ToastProvider>
  );
}

function findPatchCall() {
  return mockRequest.mock.calls.find(
    ([path, options]) => String(path).endsWith("/api/v1/users/user-1") && (options as { method?: string } | undefined)?.method === "PATCH"
  );
}

describe("ProfilePage behavior", () => {
  beforeEach(() => {
    mockRequest.mockReset();
  });

  it("normalizes the read-only birth date and allows adding social links", async () => {
    const user = userEvent.setup();
    renderProfile();

    expect(await screen.findByText("2001-02-03")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /редактировать/i }));
    await user.click(screen.getByRole("button", { name: /добавить ссылку/i }));
    await user.type(screen.getByLabelText(/платформа/i), "VK");
    await user.type(screen.getByLabelText(/ссылка/i), "https://vk.example/ivan");
    await user.click(screen.getByRole("button", { name: /сохранить/i }));

    await waitFor(() => expect(findPatchCall()).toBeTruthy());
    expect((findPatchCall()?.[1] as { body: Partial<UserProfile> }).body.social_links).toEqual([
      { platform: "VK", value: "https://vk.example/ivan" }
    ]);
  });

  it("sends an empty social_links array when the last link is removed", async () => {
    const user = userEvent.setup();
    renderProfile({
      ...baseProfile,
      social_links: [{ platform: "VK", value: "https://vk.example/ivan" }]
    });

    expect(await screen.findByRole("link", { name: "VK" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /редактировать/i }));
    await user.click(screen.getByRole("button", { name: /удалить ссылку/i }));
    await user.click(screen.getByRole("button", { name: /сохранить/i }));

    await waitFor(() => expect(findPatchCall()).toBeTruthy());
    expect((findPatchCall()?.[1] as { body: Partial<UserProfile> }).body.social_links).toEqual([]);
  });

  it("blocks saving required empty fields and shows validation feedback", async () => {
    const user = userEvent.setup();
    renderProfile();

    await screen.findByText("2001-02-03");
    await user.click(screen.getByRole("button", { name: /редактировать/i }));
    await user.clear(screen.getByDisplayValue("Иван"));
    await user.click(screen.getByRole("button", { name: /сохранить/i }));

    expect(await screen.findAllByText(/имя обязательно/i)).not.toHaveLength(0);
    expect(findPatchCall()).toBeUndefined();
  });
});
