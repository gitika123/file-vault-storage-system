import { test, expect } from '@playwright/test'

test('signed-out vault presents the secure login surface', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Your private workspace.')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Sign in to vault' })).toBeVisible()
})

test('seeded user can reach the authenticated vault', async ({ page }) => {
  const email = process.env.UAT_EMAIL
  const password = process.env.UAT_PASSWORD
  test.skip(!email || !password, 'Set UAT_EMAIL and UAT_PASSWORD for authenticated UAT.')
  await page.goto('/')
  await page.getByLabel('Email').fill(email!)
  await page.getByLabel('Password').fill(password!)
  await page.getByRole('button', { name: 'Sign in to vault' }).click()
  await expect(page.getByText('Vault protected')).toBeVisible()
  await expect(page.getByPlaceholder('Search your files')).toBeVisible()
})

test('authenticated sidebar switches between vault sections', async ({ page }) => {
  const email = process.env.UAT_EMAIL
  const password = process.env.UAT_PASSWORD
  test.skip(!email || !password, 'Set UAT_EMAIL and UAT_PASSWORD for authenticated UAT.')
  await page.goto('/')
  await page.getByLabel('Email').fill(email!)
  await page.getByLabel('Password').fill(password!)
  await page.getByRole('button', { name: 'Sign in to vault' }).click()
  await expect(page.getByRole('button', { name: 'Folders' })).toBeVisible()
  await page.getByRole('button', { name: 'Folders' }).click()
  await expect(page.getByRole('heading', { name: 'Folders' })).toBeVisible()
  await page.getByRole('button', { name: 'Uploads' }).click()
  await expect(page.getByRole('heading', { name: 'Recent uploads' })).toBeVisible()
})
