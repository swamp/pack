/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

// Package swamppack packs constants into a .swamp-pack.
package swamppack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	raff "github.com/piot/raff-go/src"
)

// ConstantType represents the type of constant stored.
type ConstantType uint8

const (
	ConstantTypeString ConstantType = iota
	ConstantTypeResourceName
	ConstantTypeInteger
	ConstantTypeBoolean
	ConstantTypeExternalFunc
	ConstantTypeFunctionDeclaration
)

type TypeRef uint16

// Function has the name, signature and opcodes fot the function.
type Function struct {
	name           string
	signature      TypeRef
	parameterCount uint
	variableCount  uint
	constants      []*Constant
	opcodes        []byte
}

// NewFunction creates a new function.
func NewFunction(name string, signature TypeRef, parameterCount uint, variableCount uint,
	constants []*Constant, opcodes []byte) *Function {
	for index, constant := range constants {
		if constant == nil {
			panic(fmt.Sprintf("you sent in bad constants at index %v %v", index, constants))
		}
	}

	return &Function{
		name: name, signature: signature, parameterCount: parameterCount,
		variableCount: variableCount, constants: constants, opcodes: opcodes,
	}
}

func (f *Function) String() string {
	return fmt.Sprintf("[fun %s signature:%v parameter:%d varcount:%d constant count:%d", f.name, f.signature,
		f.parameterCount, f.variableCount, len(f.constants))
}

// ExternalFunction represents a engine built in and external function (to the compiler).
type ExternalFunction struct {
	name           string
	signature      TypeRef
	parameterCount uint
}

// NewExternalFunction creates a new external function constant.
func NewExternalFunction(name string, parameterCount uint) *ExternalFunction {
	return &ExternalFunction{name: name, parameterCount: parameterCount}
}

func (f *ExternalFunction) String() string {
	return fmt.Sprintf("[fun %s signature:%v parameter:%d", f.name, f.signature, f.parameterCount)
}

// FunctionDeclaration holds the function declaration header.
type FunctionDeclaration struct {
	name           string
	signature      TypeRef
	parameterCount uint
}

// NewFunctionDeclaration creates a new function declaration.
func NewFunctionDeclaration(name string, signature TypeRef, parameterCount uint) *FunctionDeclaration {
	return &FunctionDeclaration{name: name, signature: signature, parameterCount: parameterCount}
}

func (f *FunctionDeclaration) String() string {
	return fmt.Sprintf("[fundeclaration %s signature:%v parameter:%d", f.name, f.signature, f.parameterCount)
}

type IndexPositionInFile = uint16

// Constant is a union of all the constant types.
type Constant struct {
	v                   int32
	boolean             bool
	indexPositionInFile IndexPositionInFile
	constantType        ConstantType
	externalFunction    *ExternalFunction
	functionDeclaration *FunctionDeclaration
	str                 string
}

func (c *Constant) String() string {
	switch c.constantType {
	case ConstantTypeBoolean:
		return fmt.Sprintf("%v", c.boolean)
	case ConstantTypeInteger:
		return fmt.Sprintf("int: %v", c.v)
	case ConstantTypeExternalFunc:
		return fmt.Sprintf("externalfunc %v", c.externalFunction)
	case ConstantTypeFunctionDeclaration:
		return fmt.Sprintf("declarefunc %v", c.functionDeclaration)
	case ConstantTypeString:
		return fmt.Sprintf("'%v'", c.str)
	case ConstantTypeResourceName:
		return fmt.Sprintf("resource name '%v'", c.str)
	}

	panic(fmt.Errorf("unknown constant type %v", c.constantType))
}

// NewStringConstant creates a new string constant.
func NewStringConstant(str string) *Constant {
	return &Constant{constantType: ConstantTypeString, str: str}
}



// NewResourceNameConstant creates a new string constant.
func NewResourceNameConstant(str string) *Constant {
	return &Constant{constantType: ConstantTypeResourceName, str: str}
}

// NewIntegerConstant creates a new integer constant.
func NewIntegerConstant(v int32) *Constant {
	return &Constant{constantType: ConstantTypeInteger, v: v}
}

// NewBooleanConstant creates a new boolean constant.
func NewBooleanConstant(b bool) *Constant {
	return &Constant{constantType: ConstantTypeBoolean, boolean: b}
}

// NewExternalFuncConstant creates a new external function constant reference.
func NewExternalFuncConstant(externalFunction *ExternalFunction) *Constant {
	return &Constant{constantType: ConstantTypeExternalFunc, externalFunction: externalFunction}
}

// NewFunctionDeclarationConstant creates a new function declaration constant.
func NewFunctionDeclarationConstant(functionDeclaration *FunctionDeclaration) *Constant {
	return &Constant{constantType: ConstantTypeFunctionDeclaration, functionDeclaration: functionDeclaration}
}

type FunctionRefIndex uint16

// ConstantRepo contains all the constants.
type ConstantRepo struct {
	stringConstants              []*Constant
	resourceNameConstants        []*Constant
	integerConstants             []*Constant
	functions                    []*Function
	externalFuncConstants        []*Constant
	functionDeclarationConstants []*Constant
	booleanConstants             []*Constant
}

// NewConstantRepo creates a new constant repo.
func NewConstantRepo() *ConstantRepo {
	return &ConstantRepo{}
}

func (s *ConstantRepo) String() string {
	return fmt.Sprintf("%v\n%v\n%v\n%v\n", s.stringConstants, s.integerConstants, s.functions, s.booleanConstants)
}

func (s *ConstantRepo) findString(str string) *Constant {
	for _, stringConstant := range s.stringConstants {
		if stringConstant.str == str {
			return stringConstant
		}
	}

	return nil
}

func (s *ConstantRepo) AddString(str string) *Constant {
	foundConstant := s.findString(str)
	if foundConstant == nil {
		foundConstant = NewStringConstant(str)
		s.stringConstants = append(s.stringConstants, foundConstant)
	}

	return foundConstant
}



func (s *ConstantRepo) findResourceName(str string) *Constant {
	for _, resourceNameConstant := range s.resourceNameConstants {
		if resourceNameConstant.str == str {
			return resourceNameConstant
		}
	}

	return nil
}

func (s *ConstantRepo) AddResourceName(str string) *Constant {
	foundConstant := s.findResourceName(str)
	if foundConstant == nil {
		foundConstant = NewResourceNameConstant(str)
		s.resourceNameConstants = append(s.resourceNameConstants, foundConstant)
	}

	return foundConstant
}


func (s *ConstantRepo) findInteger(v int32) *Constant {
	for _, integerConstant := range s.integerConstants {
		if integerConstant.v == v {
			return integerConstant
		}
	}

	return nil
}

func (s *ConstantRepo) AddInteger(v int32) *Constant {
	foundConstant := s.findInteger(v)
	if foundConstant == nil {
		foundConstant = NewIntegerConstant(v)
		s.integerConstants = append(s.integerConstants, foundConstant)
	}

	return foundConstant
}

func (s *ConstantRepo) findBoolean(b bool) *Constant {
	for _, booleanConstant := range s.booleanConstants {
		if booleanConstant.boolean == b {
			return booleanConstant
		}
	}

	return nil
}

func (s *ConstantRepo) AddBoolean(b bool) *Constant {
	foundConstant := s.findBoolean(b)
	if foundConstant == nil {
		foundConstant = NewBooleanConstant(b)
		s.booleanConstants = append(s.booleanConstants, foundConstant)
	}

	return foundConstant
}

func (s *ConstantRepo) AddFunctionReference(uniqueFullyQualifiedName string) (*Constant, error) {
	foundConstant := s.FindFunctionDeclaration(uniqueFullyQualifiedName)
	if foundConstant == nil {
		return nil, fmt.Errorf("couldn't find a previous declaration for '%v' and that is required", uniqueFullyQualifiedName)
	}

	return foundConstant, nil
}

func (s *ConstantRepo) AddExternalFunctionReference(uniqueFullyQualifiedName string) (*Constant, error) {
	foundConstant := s.FindExternalFunction(uniqueFullyQualifiedName)
	if foundConstant == nil {
		return nil, fmt.Errorf("couldn't find a previous external function declaration for '%v' and that is required", uniqueFullyQualifiedName)
	}

	return foundConstant, nil
}

func (s *ConstantRepo) FindExternalFunction(name string) *Constant {
	for _, externalFuncConstant := range s.externalFuncConstants {
		if externalFuncConstant.externalFunction.name == name {
			return externalFuncConstant
		}
	}

	return nil
}

func (s *ConstantRepo) FindFunctionDeclaration(name string) *Constant {
	for _, functionDeclarationConstant := range s.functionDeclarationConstants {
		if functionDeclarationConstant.functionDeclaration.name == name {
			return functionDeclarationConstant
		}
	}

	return nil
}

func (s *ConstantRepo) FindFunctionDeclarationByIndex(index FunctionRefIndex) *Constant {
	return s.functionDeclarationConstants[index]
}

func (s *ConstantRepo) AddFunction(name string, signature TypeRef, parameterCount uint, variableCount uint,
	constants []*Constant, opcodes []byte) *Function {
	f := NewFunction(name, signature, parameterCount, variableCount, constants, opcodes)
	s.functions = append(s.functions, f)

	return f
}

func (s *ConstantRepo) AddExternalFunction(name string, parameterCount uint) *Constant {
	foundExternalFuncConst := s.FindExternalFunction(name)
	if foundExternalFuncConst == nil {
		f := NewExternalFunction(name, parameterCount)
		foundExternalFuncConst = NewExternalFuncConstant(f)
		s.externalFuncConstants = append(s.externalFuncConstants, foundExternalFuncConst)
	}

	return foundExternalFuncConst
}

func (s *ConstantRepo) AddFunctionDeclaration(name string, signature TypeRef, parameterCount uint) *Constant {
	foundFunctionDeclarationConst := s.FindFunctionDeclaration(name)
	if foundFunctionDeclarationConst == nil {
		f := NewFunctionDeclaration(name, signature, parameterCount)
		foundFunctionDeclarationConst = NewFunctionDeclarationConstant(f)
		s.functionDeclarationConstants = append(s.functionDeclarationConstants, foundFunctionDeclarationConst)
	}

	return foundFunctionDeclarationConst
}

func writeBools(booleanConstants []*Constant, writer io.Writer, indexOffset int) (int, error) {
	count := len(booleanConstants)
	header := []byte{byte(count)}

	booleanIcon := raff.FourOctets{0xF0, 0x9F, 0x90, 0x9C}
	if err := raff.WriteInternalChunkMarker(writer, booleanIcon); err != nil {
		return -1, err
	}

	if _, writeErr := writer.Write(header); writeErr != nil {
		return 0, writeErr
	}

	octets := make([]byte, count)

	for index, b := range booleanConstants {
		if b.constantType != ConstantTypeBoolean {
			panic("wrong boolean type")
		}

		valueToWrite := uint8(0)

		if b.boolean {
			valueToWrite = uint8(1)
		}

		b.indexPositionInFile = IndexPositionInFile(indexOffset + index)

		octets[index] = valueToWrite
	}

	if _, err := writer.Write(octets); err != nil {
		return 0, err
	}

	return count, nil
}

func writeIntegers(integerConstants []*Constant, writer io.Writer, indexOffset int) (int, error) {
	count := len(integerConstants)

	header := []byte{byte(count)}

	integerIcon := raff.FourOctets{0xF0, 0x9F, 0x94, 0xA2}
	if err := raff.WriteInternalChunkMarker(writer, integerIcon); err != nil {
		return -1, err
	}

	if _, writeErr := writer.Write(header); writeErr != nil {
		panic(writeErr)
	}

	for index, constant := range integerConstants {
		if constant.constantType != ConstantTypeInteger {
			panic("wrong integer type")
		}

		constant.indexPositionInFile = IndexPositionInFile(indexOffset + index)

		if writeErr := binary.Write(writer, binary.BigEndian, constant.v); writeErr != nil {
			return 0, writeErr
		}
	}

	return count, nil
}

func writeString(str string, writer io.Writer) error {
	stringOctets := []byte(str)

	if _, writeErr := writer.Write([]byte{byte(len(str))}); writeErr != nil {
		return writeErr
	}

	if _, writeErr := writer.Write(stringOctets); writeErr != nil {
		return writeErr
	}

	return nil
}

func writeTypeRef(typeRef TypeRef, writer io.Writer) error {
	_, err := writer.Write([]byte{byte(typeRef)})

	return err
}

func writeStrings(stringConstants []*Constant, writer io.Writer, indexOffset int) (int, error) {
	count := len(stringConstants)
	header := []byte{byte(count)}

	stringsIcon := raff.FourOctets{0xF0, 0x9F, 0x8E, 0xBB}
	if err := raff.WriteInternalChunkMarker(writer, stringsIcon); err != nil {
		return -1, err
	}

	if _, writeErr := writer.Write(header); writeErr != nil {
		return 0, writeErr
	}

	for index, constant := range stringConstants {
		if constant.constantType != ConstantTypeString {
			panic("wrong string type")
		}

		constant.indexPositionInFile = IndexPositionInFile(indexOffset + index)

		if writeErr := writeString(constant.str, writer); writeErr != nil {
			return 0, writeErr
		}
	}

	return count, nil
}

func writeResourceNames(resourceNameConstants []*Constant, writer io.Writer, indexOffset int) (int, error) {
	count := len(resourceNameConstants)
	header := []byte{byte(count)}

	resourceNameIcon := raff.FourOctets{0xF0, 0x9F, 0x8C, 0xB3}
	if err := raff.WriteInternalChunkMarker(writer, resourceNameIcon); err != nil {
		return -1, err
	}

	if _, writeErr := writer.Write(header); writeErr != nil {
		return 0, writeErr
	}

	for index, constant := range resourceNameConstants {
		if constant.constantType != ConstantTypeResourceName {
			panic("wrong resourceType type")
		}

		constant.indexPositionInFile = IndexPositionInFile(indexOffset + index)

		if writeErr := writeString(constant.str, writer); writeErr != nil {
			return 0, writeErr
		}
	}

	return count, nil
}

func writeExternalFunctions(externalFuncConstants []*Constant, writer io.Writer, indexOffset int) (int, error) {
	count := len(externalFuncConstants)
	header := []byte{byte(count)}

	externalFunctionIcons := raff.FourOctets{0xF0, 0x9F, 0x91, 0xBE}
	if err := raff.WriteInternalChunkMarker(writer, externalFunctionIcons); err != nil {
		return -1, err
	}

	if _, writeErr := writer.Write(header); writeErr != nil {
		return 0, writeErr
	}

	for index, constant := range externalFuncConstants {
		if constant.constantType != ConstantTypeExternalFunc {
			panic("external_func: wrong func type")
		}

		constant.indexPositionInFile = IndexPositionInFile(indexOffset + index)

		f := constant.externalFunction
		header := []byte{byte(f.parameterCount)}

		if _, writeErr := writer.Write(header); writeErr != nil {
			return 0, writeErr
		}

		if writeErr := writeString(f.name, writer); writeErr != nil {
			return 0, writeErr
		}

		if writeErr := writeTypeRef(f.signature, writer); writeErr != nil {
			return 0, writeErr
		}
	}

	return count, nil
}

func writeFunctionDeclarations(externalFuncConstants []*Constant, writer io.Writer, indexOffset int) (int, error) {
	count := len(externalFuncConstants)
	header := []byte{0, 0, 0, 0}

	functionDeclarationIcon := raff.FourOctets{0xF0, 0x9F, 0x9B, 0x82}
	if err := raff.WriteInternalChunkMarker(writer, functionDeclarationIcon); err != nil {
		return -1, err
	}

	binary.BigEndian.PutUint32(header[0:], uint32(count))

	if _, writeErr := writer.Write(header); writeErr != nil {
		return 0, writeErr
	}

	for index, constant := range externalFuncConstants {
		if constant.constantType != ConstantTypeFunctionDeclaration {
			panic("external_func: wrong func type")
		}

		constant.indexPositionInFile = IndexPositionInFile(indexOffset + index)

		f := constant.functionDeclaration

		header := []byte{byte(f.parameterCount)}
		if _, writeErr := writer.Write(header); writeErr != nil {
			return 0, writeErr
		}

		if writeErr := writeString(f.name, writer); writeErr != nil {
			return 0, writeErr
		}

		if writeErr := writeTypeRef(f.signature, writer); writeErr != nil {
			return 0, writeErr
		}
	}

	return count, nil
}

func writeFunctions(functions []*Function, writer io.Writer) (int, error) {
	count := len(functions)
	header := []byte{0, 0, 0, 0}

	binary.BigEndian.PutUint32(header[0:], uint32(count))

	functionIcon := raff.FourOctets{0xF0, 0x9F, 0x90, 0x8A}
	if err := raff.WriteInternalChunkMarker(writer, functionIcon); err != nil {
		return -1, err
	}

	if _, writeErr := writer.Write(header); writeErr != nil {
		return 0, writeErr
	}

	for _, f := range functions {
		header := []byte{
			byte(f.parameterCount), byte(f.variableCount), byte(0),
			byte(len(f.constants)),
		} // was temp count

		if _, writeErr := writer.Write(header); writeErr != nil {
			return 0, writeErr
		}

		for _, subConstant := range f.constants {
			if subConstant == nil {
				panic(fmt.Sprintf("how can subconstant be nil? %v", f.constants))
			}

			indexInFile := subConstant.indexPositionInFile

			const NotSetPosition IndexPositionInFile = 0xffff

			if indexInFile == NotSetPosition {
				panic(fmt.Errorf("wrong index for constant %v in function %v", subConstant, f))
			}
			constantIndexBigEndian := []byte{0, 0}

			binary.BigEndian.PutUint16(constantIndexBigEndian, indexInFile)
			if _, writeErr := writer.Write(constantIndexBigEndian); writeErr != nil {
				return 0, writeErr
			}
		}

		opcodeCountHeader := []byte{0, 0}

		binary.BigEndian.PutUint16(opcodeCountHeader, uint16(len(f.opcodes)))

		if _, writeErr := writer.Write(opcodeCountHeader); writeErr != nil {
			return 0, writeErr
		}

		if _, writeErr := writer.Write(f.opcodes); writeErr != nil {
			return 0, writeErr
		}
	}

	return count, nil
}

type Version struct {
	Major uint8
	Minor uint8
	Patch uint8
}

func writeChunkHeader(writer io.Writer, icon raff.FourOctets, name raff.FourOctets, payload []byte) error {
	if err := raff.WriteChunk(writer, icon, name, payload); err != nil {
		return err
	}

	return nil
}

func writePackHeader(writer io.Writer) error {
	name := raff.FourOctets{'s', 'p', 'k', '4'}
	packetIcon := raff.FourOctets{0xF0, 0x9F, 0x93, 0xA6}
	return writeChunkHeader(writer, packetIcon, name, nil)
}

func writeVersion(writer io.Writer, version Version) error {
	header := []byte{version.Major, version.Minor, version.Patch}

	if _, writeErr := writer.Write(header); writeErr != nil {
		return writeErr
	}

	return nil
}

func writeTypeInfo(writer io.Writer, payload []byte) error {
	name := raff.FourOctets{'s', 't', 'i', '0'}
	packetIcon := raff.FourOctets{0xF0, 0x9F, 0x93, 0x9C}
	return writeChunkHeader(writer, packetIcon, name, payload)
}

func writeCodeChunk(writer io.Writer, payload []byte) error {
	name := raff.FourOctets{'s', 'c', 'd', '0'}
	packetIcon := raff.FourOctets{0xF0, 0x9F, 0x92, 0xBB}
	return writeChunkHeader(writer, packetIcon, name, payload)
}

func packCode(constants *ConstantRepo) ([]byte, error) {
	var err error
	var buf bytes.Buffer

	indexOffset := 0
	offset := 0

	offset, err = writeExternalFunctions(constants.externalFuncConstants, &buf, indexOffset)
	if err != nil {
		return nil, err
	}

	indexOffset += offset

	offset, err = writeFunctionDeclarations(constants.functionDeclarationConstants, &buf, indexOffset)
	if err != nil {
		return nil, err
	}

	indexOffset += offset


	offset, err = writeBools(constants.booleanConstants, &buf, indexOffset)
	if err != nil {
		return nil, err
	}

	indexOffset += offset

	offset, err = writeIntegers(constants.integerConstants, &buf, indexOffset)
	if err != nil {
		return nil, err
	}

	indexOffset += offset

	offset, err = writeStrings(constants.stringConstants, &buf, indexOffset)
	if err != nil {
		return nil, err
	}

	indexOffset += offset

	if _, err := writeResourceNames(constants.resourceNameConstants, &buf, indexOffset); err != nil {
		return nil, err
	}

	if _, err := writeFunctions(constants.functions, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeCode(writer io.Writer, constants *ConstantRepo) error {
	payload, codeErr := packCode(constants)
	if codeErr != nil {
		return codeErr
	}

	return writeCodeChunk(writer, payload)
}

// Pack writes all constants to a .swamp-pack file.
func Pack(constants *ConstantRepo, typeInfo []byte) ([]byte, error) {
	var buf bytes.Buffer

	if err := raff.WriteHeader(&buf); err != nil {
		return nil, err
	}

	if writeErr := writePackHeader(&buf); writeErr != nil {
		return nil, writeErr
	}

	if writeErr := writeTypeInfo(&buf, typeInfo); writeErr != nil {
		return nil, writeErr
	}

	if writeErr := writeCode(&buf, constants); writeErr != nil {
		return nil, writeErr
	}

	return buf.Bytes(), nil
}
