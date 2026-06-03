import { api } from './api'

export interface Province {
  name: string
  code: number
  codename: string
}

export interface District {
  name: string
  code: number
  codename: string
}

export interface Ward {
  name: string
  code: number
  codename: string
}

export const locationAPI = {
  getProvinces: async (): Promise<Province[]> => {
    const res = await api.get<Province[]>('/location/provinces')
    return res.data || []
  },

  getDistricts: async (provinceCode: number): Promise<District[]> => {
    const res = await api.get<District[]>(`/location/provinces/${provinceCode}/districts`)
    return res.data || []
  },

  getWards: async (districtCode: number): Promise<Ward[]> => {
    const res = await api.get<Ward[]>(`/location/districts/${districtCode}/wards`)
    return res.data || []
  },
}
