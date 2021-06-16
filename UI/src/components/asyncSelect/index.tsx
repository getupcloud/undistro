import React, { FC } from 'react'
import { AsyncPaginate } from "react-select-async-paginate";

export type OptionType = { value: string, label: string }

type LoadOptionResult = {
  options: OptionType[];
  hasMore: boolean;
  additional: { page: number };
}

type Props = {
  label?: string,
  onChange: (option: OptionType | null) => void,
  loadOptions: any,
  value: OptionType | null
}

const AsyncSelect: FC<Props> = ({ 
  label,
  onChange,
  loadOptions,
  value
}) => {

  const handleChange = (option: any) => {
    onChange(option.value)
  }

  return (
    <div className='select'>
    <div className='title-section'>
      <label>{label}</label>
    </div>

    <AsyncPaginate
      defaultOptions
      loadOptions={loadOptions}
      onChange={(e) => handleChange(e)}
      classNamePrefix="select-container"
      value={value}
      additional={{ page: 1 }}
    />
  </div>
  )
}

export default AsyncSelect