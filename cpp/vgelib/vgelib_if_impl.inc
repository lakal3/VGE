
thread_local std::string lastError;

DLLEXPORT int Get_Error(void*& str, int64_t& len)
{
	len = lastError.size();
	str = reinterpret_cast<void*>(lastError.data());
	return 0;
}


