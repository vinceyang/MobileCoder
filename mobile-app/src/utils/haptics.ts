import { Haptics, ImpactStyle } from '@capacitor/haptics'

export async function lightImpact(): Promise<void> {
  try {
    await Haptics.impact({ style: ImpactStyle.Light })
  } catch {
    // Haptics not available on this device
  }
}

export async function mediumImpact(): Promise<void> {
  try {
    await Haptics.impact({ style: ImpactStyle.Medium })
  } catch {
    // Haptics not available on this device
  }
}

export async function heavyImpact(): Promise<void> {
  try {
    await Haptics.impact({ style: ImpactStyle.Heavy })
  } catch {
    // Haptics not available on this device
  }
}
