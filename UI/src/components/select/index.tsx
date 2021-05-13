import React, { FC } from 'react'
import Select from 'react-select'

import './index.scss'

const optionsDefault = [
  { value: 'chocolate', label: 'Chocolate' },
  { value: 'strawberry', label: 'Strawberry' },
  { value: 'vanilla', label: 'Vanilla' }
]

type options = [
  { value: any, label: string }
]

type Props = {
  label?: string,
  options?: any,
  onChange?: any,
  value?: any
}

const SelectUndistro: FC<Props> = ({ 
  label,
  options,
  onChange,
  value
}) => {
  return (
    <div className='select'>
    <div className='title-section'>
      <label>{label}</label>
    </div>

    <Select
      options={options}
      onChange={onChange}
      value={value}
      classNamePrefix="select-container"
    />
  </div>
  )
}

export default SelectUndistro