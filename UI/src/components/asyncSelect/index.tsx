import React, { FC } from 'react'
import { TypeAsyncSelect } from '../../types/cluster'
import { AsyncPaginate } from "react-select-async-paginate";

const AsyncSelect: FC<TypeAsyncSelect> = ({ 
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