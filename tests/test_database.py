import pytest


@pytest.mark.asyncio
async def test_create_connection_without_error(db):
    async with db.cursor():
        pass


@pytest.mark.asyncio
async def test_check_database_exists_when_database_does_not_exist(db):
    assert not await db.database_exists("i_dont_exist")


@pytest.mark.asyncio
async def test_check_database_exists_when_database_already_exists(db):
    assert await db.database_exists("database_exists")


@pytest.mark.asyncio
async def test_check_user_exists_when_user_does_not_exist(db):
    assert not await db.user_exists("i_dont_exist")


@pytest.mark.asyncio
async def test_check_user_exists_when_user_already_exists(db):
    assert await db.user_exists("user_exists")


@pytest.mark.asyncio
async def test_check_user_has_all_privileges_when_user_has_all_privileges(db):
    assert await db.user_has_all_privileges_on_database("everything_exists")


@pytest.mark.asyncio
async def test_check_user_has_all_privileges_when_user_no_privileges(db):
    assert not await db.user_has_all_privileges_on_database("user_and_db_but_no_privileged")


@pytest.mark.asyncio
async def test_create_database(db):
    name = "new_database"
    await db.create_database(name)
    assert await db.database_exists(name)


@pytest.mark.asyncio
async def test_create_user(db):
    name = "new_database"
    password = "test"
    await db.create_user(name, password)
    assert await db.user_exists(name)


@pytest.mark.asyncio
async def test_grant_privileges(db):
    name = "new_database"
    password = "test"
    await db.create_user(name, password)
    await db.create_database(name)
    await db.grant_all_privileges(name)
    assert await db.user_has_all_privileges_on_database(name)


@pytest.mark.asyncio
async def test_change_password(db):
    name = "new_database"
    password = "test"
    await db.create_user(name, password)
    await db.update_user_password(name, password)
    # afaik we would need a super user to actually assert that the password changed
    # i.e. compare SELECT passwd FROM pg_shadow WHERE usename = 'test'; before and after
